# -*- coding: utf-8 -*-
import sys
import csv
import nltk.data
import nltk.corpus
from select_aliases import create_candidate_list
from vocabulary import get_stems

reload(sys)
sys.setdefaultencoding('utf-8')

ENGLISH_STOPWORDS = set(nltk.corpus.stopwords.words('english'))


def main():
    vocabulary = {}
    with open('VOCABULARY_5000.txt', 'rb') as f:
        idx = 0
        for word in f:
            w = word.replace('\n', '')
            vocabulary[w] = idx
            idx += 1

    aliases = set()
    with open('ALIASES.txt', 'rb') as f:
        for alias in f:
            a = alias.lower().replace('\n', '')
            aliases.add(a)

    writer = csv.writer(sys.stdout)

    count = 0
    # Read SENTENCES.csv
    for row in csv.reader(sys.stdin):
        count += 1
        if count % 100 == 0:
            sys.stderr.write("%d\n" % count)
        try:
            sent_id = row[0]
            sentence = row[5]
            candidates = create_candidate_list(sentence)
            n = -1
            for candidate in candidates:
                if candidate.lower() in aliases:
                    n += 1
                    id = sent_id + '__%d' % n

                    # Make "bag of words" vector
                    bag_of_words = [0] * len(vocabulary)
                    stems = get_stems(sentence)
                    c = 0
                    for stem in stems:
                        i = vocabulary.get(stem, None)
                        if i is not None:
                            bag_of_words[i] = 1
                            c += 1

                    if c > 0:
                        # Write output.
                        writer.writerow(
                            [id, sent_id, candidate] + bag_of_words)
        except UnicodeDecodeError:
            pass
        except IndexError:
            pass

if __name__ == "__main__":
    main()
