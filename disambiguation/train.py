# -*- coding: utf-8 -*-
import sys
import re
import copy
import csv
import random
import glob
import fileinput
import cPickle as pickle
import pyley
import pylru
from sklearn.linear_model import SGDClassifier
from select_aliases import uppercase_first_letters

reload(sys)
sys.setdefaultencoding('utf-8')

csv.field_size_limit(sys.maxsize)

COLUMNS = 5000

MATRIX_KEY = "M_%s"
VECTOR_KEY = "V_%s_EPOCH_%d"
CLASSIFIER_KEY = "C_%s_EPOCH_%d"
SAMPLE_KEY = "S_%s_EPOCH_%d"
FILE_PATTERN = "/data/pickles/{key}__{version}.pickle"


def format_alias(alias):
    # Filter numbers.
    if len(alias.split()) == 1:
        try:
            float(alias)
            return None
        except ValueError:
            pass

    # uppercase first letters and return.
    return uppercase_first_letters(alias)


class Graph(object):
    client = pyley.CayleyClient()
    graph = pyley.GraphObject()

    cache = pylru.lrucache(100)  # LRU cache.

    NAME = "http://www.w3.org/2000/01/rdf-schema#label"
    ALIAS = "http://rdf.basekb.com/ns/common.topic.alias"
    ALIAS_PATTERN = "\"%s\"@en"
    TYPE = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
    # TYPE_PATTERN = "http://rdf.basekb.com/ns/%s"

    def get_entities(self, alias):
        """
        Return the list of entity ids with given alias.
        """
        # Hit the cache first.
        try:
            r = self.cache[alias]
            return r
        except KeyError:
            pass

        entities = []
        name = self.ALIAS_PATTERN % alias
        ## Query the graph:
        # Look for entities with name == alias
        q1 = self.graph.V().Has(self.NAME, name).All()
        response1 = self.client.Send(q1)
        # Look for entities with alias == alias
        q2 = self.graph.V().Has(self.ALIAS, name).All()
        response2 = self.client.Send(q2)

        ## Merge both responses to get the list of entities.
        seen_entities = set()
        if response1.result.get('result'):
            l1 = response1.result.get('result')
            for el in l1:
                e = el.get('id')
                if e not in seen_entities:
                    entities.append(e)
                    seen_entities.add(e)
        if response2.result.get('result'):
            l2 = response2.result.get('result')
            for el in l2:
                e = el.get('id')
                if e not in seen_entities:
                    entities.append(e)
                    seen_entities.add(e)

        # Cache results.
        self.cache[alias] = copy.copy(entities)

        return entities

    def get_types(self, entity):
        # Check for cached value.
        try:
            cached_result = self.cache[entity]
            return cached_result
        except KeyError:
            pass

        t = []
        q = self.graph.V(entity).Out(self.TYPE).All()
        response = self.client.Send(q)
        res = response.result.get('result')
        if not res:
            self.cache[entity] = []  # Update cache.
            return []
        seen_types = set()
        for el in res:
            e = el.get('id')
            if e not in seen_types:
                t.append(e)
                seen_types.add(e)

        self.cache[entity] = copy.copy(t)  # Update cache.
        return t

    def get_all_types(self, entities):
        """
        For each entities, return the list of unique types.
        e.g:
        ["entity_1", "entity_2"] -> [["type_b", "type_a"], ["type_c", "type_a"]]
        """
        types = []
        for entity in entities:
            types.append(self.get_types(entity))
        return types


class Loader(object):

    graph = Graph()
    X_STORE = {}
    Y_STORE = {}

    def load_data(self):
        """
        Load data to pickle files.
        """
        row_number = 0
        previous_alias = ""

        for row in csv.reader(sys.stdin):
            row_number += 1
            if row_number < 850000:
                continue
            if row_number % 100 == 0:
                sys.stderr.write("Row %d...\n" % row_number)
            alias = format_alias(row[2])
            if not alias:
                continue
            x = []
            # Fill vector x.
            for i in xrange(COLUMNS):
                if i < 3:
                    continue
                x.append(int(row[i]))

            if alias != previous_alias:
                # Save previous values to disk.
                self.save_to_disk()
                previous_alias = alias

            entities = self.graph.get_entities(alias)
            types = self.graph.get_types(entities)

            idx = random.randint(0, len(entities))  # Assign random entity.
            i = -1
            for entity in entities:
                entity_id = entity.split('/')[-1]
                i += 1
                # Assign alias to one of the entities.
                y = 0
                if i == idx:
                    y = 1

                # Update entities.
                self.update(entity_id, x, y)

                # Update types.
                if not types:
                    continue
                for t in types[i]:
                    type_id = t.split('/')[-1]
                    self.update(type_id, x, y)

    def update(self, uid, x, y):
        X_key = MATRIX_KEY % uid
        Y_key = VECTOR_KEY % (uid, 0)  # Epoch 0.

        # Get previous X and Y
        X = self.X_STORE.get(X_key) or []
        Y = self.Y_STORE.get(Y_key) or []

        # Append new values.
        X.append(x)
        Y.append(y)

        # Store X and Y back.
        self.X_STORE[X_key] = X
        self.Y_STORE[Y_key] = Y

    def save_to_disk(self):
        # Save X and Y to disk.
        for STORE in [self.X_STORE, self.Y_STORE]:
            for key, X in STORE.iteritems():
                filepattern = FILE_PATTERN.format(key=key, version="*")

                files = glob.glob(filepattern)
                version = len(files) + 1
                filepath = FILE_PATTERN.format(
                    key=key,
                    version=str(version),
                )

                with open(filepath, 'w') as f:
                    # override previous pickle
                    pickle.dump(X, f, -1)

        # Clear X and Y stores.
        self.X_STORE = {}
        self.Y_STORE = {}


class Classifier(object):

    graph = Graph()

    Y_STORE = {}

    entity_id_regexp = re.compile(r'^.+/([^>]+)>?$')
    entity_regexp = re.compile(r'^<?([^<>]+/[^<>]+)>?$')

    MAX_SAMPLES = 50000

    PROBABILITY_THRESHOLD = 0.0

    def __init__(self, epoch=0, alpha=0.1):
        self.epoch = epoch
        self.alpha = alpha  # Weight of the sum of P(y|t), for t € T(e).

    def train(self):
        # Read ENTITIES.txt or TYPES.txt from stdin.
        l = 0
        for line in fileinput.input():
            l += 1

            # if l < 14912:
            #     continue

            entity_id = self.entity_id_regexp.match(line).group(1)
            entity = self.entity_regexp.match(line).group(1)

            sys.stderr.write("%d - Handling %s\n" % (l, entity))

            # Get vector and matrix.
            Xall, yall = self.merge_files(entity_id, self.epoch)

            if not yall or len(yall) < 100:
                # Not enough data.
                sys.stderr.write("Ignoring %s (%d)\n" % (entity, len(yall)))
                continue

            sys.stderr.write("Length: %d\n" % len(yall))

            # Fit entity.
            self.classify(entity_id, Xall, yall, self.epoch)

    def merge_files(self, entity_id):
        """
        Merge pickle files together to re-generate X and y.
        """
        Xall = self.get_X(entity_id)
        yall = self.get_y(entity_id)

        return Xall, yall

    def balance_dataset(self, Xall, yall, key):
        """
        Split the dataset in two parts and balance one of the part
        between valid and invalid elements.
        """
        l = len(yall)

        # Select 2/3 of the indexes randomly.
        nb_samples = 2 * (l/3)
        sample = set(random.sample(xrange(l), nb_samples))

        # # store sampling.
        # k = SAMPLE_KEY % (key, self.epoch)
        # filepath = FILE_PATTERN.format(
        #     key=k,
        #     version="",  # Whatever
        # )
        # with open(filepath, 'w') as f:
        #     pickle.dump(sample, f, -1)

        # Count the number of valid instances.
        valid_idx = []
        invalid_idx = []
        idx = -1
        for instance in yall:
            idx += 1
            if idx not in sample:
                continue
            if instance == 0:
                # invalid
                invalid_idx.append(idx)
            else:
                # valid
                valid_idx.append(idx)

        if len(valid_idx) == 0 or len(invalid_idx) == 0:
            sys.stderr.write("Only one class labels!")
            return [], []

        s = set()
        ms = self.MAX_SAMPLES / 2
        if len(valid_idx) < len(invalid_idx):
            m = len(valid_idx) if len(valid_idx) <= ms else ms
            v = valid_idx
            if len(valid_idx) > ms:
                v = random.sample(valid_idx, m)
            li = random.sample(invalid_idx, m) + v
            s = set(li)
        else:
            m = len(invalid_idx) if len(invalid_idx) <= ms else ms
            v = invalid_idx
            if len(invalid_idx) > ms:
                v = random.sample(invalid_idx, m)
            li = random.sample(valid_idx, m) + v
            s = set(li)

        X, y = [], []
        idx = -1
        for instance in yall:
            idx += 1
            if idx not in sample:
                continue

            if idx in s:
                X.append(Xall[idx])
                y.append(yall[idx])
        return X, y

    def classify(self, entity_id, Xall, yall):
        # Balance the dataset.
        X, y = self.balance_dataset(Xall, yall, entity_id, self.epoch)

        if not y:
            return

        sys.stderr.write("After balancing: %d\n" % len(y))

        # Fit the classifier.
        clf = SGDClassifier(loss="log", shuffle=True).fit(X, y)

        # Store the classifier.
        c_key = CLASSIFIER_KEY % (entity_id, self.epoch)
        c_file = FILE_PATTERN.format(key=c_key, version="")
        with open(c_file, 'w') as f:
            pickle.dump(clf, f, -1)

    def assign(self, alias, f):
        """
        Given an alias and a bag-of-words vector, assign an entity.
        Return the id bag-of-words the selected entity or None.
        """
        # get entities
        entities = self.graph.get_entities(alias)

        if len(entities) == 1:
            # Naturally disambiguated alias.
            return entities[0], 1.0

        elif not entities:
            return None, .0

        # get all types
        all_types = self.graph.get_all_types(entities)

        probabilities = {}

        idx = -1
        for entity in entities:
            entity_id = entity.split('/')[-1]
            idx += 1

            clf = None
            c_key = CLASSIFIER_KEY % (entity_id, self.epoch)
            c_file = FILE_PATTERN.format(key=c_key, version="")
            try:
                with open(c_file, 'r') as fi:
                    clf = pickle.load(fi)
            except IOError:
                return None, .0

            p = 0
            if clf is not None:
                i = 0 if clf.classes_[0] == 1 else 1
                p = clf.predict_proba(f)[0][i]

            types = all_types[idx]
            for t in types:
                clf_t = None
                type_id = t.split('/')[-1]
                c_key = CLASSIFIER_KEY % (type_id, self.epoch)
                c_file = FILE_PATTERN.format(key=c_key, version="")
                try:
                    with open(c_file, 'r') as fi:
                        clf_t = pickle.load(fi)
                except IOError:
                    pass

                if clf_t is not None:
                    i = 0 if clf_t.classes_[0] == 1 else 1
                    p += self.alpha * clf_t.predict_proba(f)[0][i]

            probabilities[entity] = p

        # Assign entity!
        max_proba = 0
        selected_entity = None
        for k, v in probabilities.iteritems():
            if v > max_proba:
                max_proba = v
                selected_entity = k

        if max_proba > self.PROBABILITY_THRESHOLD:
        # if max_proba > 0:
            return selected_entity, max_proba
        else:
            return None, .0

    def assign_all(self):
        writer = csv.writer(sys.stdout)
        row_number = 0
        assignment_nb = 0
        #previous_alias = ""
        for row in csv.reader(sys.stdin):
            row_number += 1
            if row_number % 1000 == 0:
                sys.stderr.write("Row %d...\n" % row_number)

            if row_number < 780433:
                continue

            alias = format_alias(row[2])
            if not alias:
                continue
            x = []
            # Fill vector x.
            for i in xrange(COLUMNS):
                if i < 3:
                    continue
                x.append(int(row[i]))

            # if alias != previous_alias:
            #     # Save previous values to disk.
            #     self.save_to_disk()
            #     previous_alias = alias

            # Assign entity.
            entity, p = self.assign(alias, x)

            if entity:
                assignment_nb += 1
                sys.stderr.write("Assignment %d / %d (Proba: %f)\n" % (
                    assignment_nb, row_number, p))
                writer.writerow([row[0], row[1], alias, entity, p])

            # # get entities
            # entities = self.graph.get_entities(alias)
            # # get all types
            # all_types = self.graph.get_all_types(entities)

            # if not entity:
            #     self.update_unchanged(entities, all_types, row_number)
            # else:
            #     self.update(entity, entities, all_types, row_number)

    def save_to_disk(self):
        # Save X and Y to disk.
        for key, y in self.Y_STORE.iteritems():
            filepattern = FILE_PATTERN.format(key=key, version="*")

            files = glob.glob(filepattern)
            version = len(files) + 1
            filepath = FILE_PATTERN.format(
                key=key,
                version=str(version),
            )

            with open(filepath, 'w') as f:
                # override previous pickle
                pickle.dump(y, f, -1)

        # Clear store.
        self.Y_STORE = {}

    def get_X(self, entity_id):
        m_key = MATRIX_KEY % entity_id
        m_file = FILE_PATTERN.format(key=m_key, version="*")
        X = []
        for fp in sorted(glob.glob(m_file)):
            # Put back all parts together.
            if len(X) >= self.MAX_SAMPLES:
                break
            with open(fp, 'r') as f:
                X += pickle.load(f)
        return X

    def get_y(self, entity_id):
        v_key = VECTOR_KEY % (entity_id, self.epoch)
        v_file = FILE_PATTERN.format(key=v_key, version="*")
        y = []
        for fp in sorted(glob.glob(v_file)):
            # Put back all parts together.
            if len(y) >= self.MAX_SAMPLES:
                break
            with open(fp, 'r') as f:
                y += pickle.load(f)
        return y

    def update_unchanged(self, entities, all_types, row_number):
        idx = -1
        for entity in entities:
            idx += 1
            e_id = entity.split('/')[-1]

            y = self.get_y(e_id)

            self.Y_STORE[e_id] = y

            for t in all_types[idx]:
                type_id = t.split('/')[-1]
                y_t = self.get_y(e_id)
                self.Y_STORE[type_id] = y_t

    def update(self, entity_id, entities, all_types, row_number):
        idx = -1
        for entity in entities:
            idx += 1
            e_id = entity.split('/')[-1]

            y = self.get_y(e_id)

            v = 0
            if entity_id == e_id:
                v = 1

            y[row_number] = v

            self.Y_STORE[e_id] = y

            for t in all_types[idx]:
                type_id = t.split('/')[-1]
                y_t = self.get_y(e_id)
                y_t[row_number] = v
                self.Y_STORE[type_id] = y_t


def main():
    # Load data.
    #loader = Loader()
    #loader.load_data()

    # Classify.
    classifier = Classifier(epoch=6)
    #classifier.train()
    classifier.assign_all()


if __name__ == "__main__":
    main()
