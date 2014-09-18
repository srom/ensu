# -*- coding: utf-8 -*-
import sys
import csv
import pyley  # https://github.com/ziyasal/pyley

reload(sys)
sys.setdefaultencoding('utf-8')

csv.field_size_limit(sys.maxsize)

NAME = "http://www.w3.org/2000/01/rdf-schema#label"
ALIAS = "http://rdf.basekb.com/ns/common.topic.alias"
ALIAS_PATTERN = "\"%s\"@en"
TYPE = "http://www.w3.org/1999/02/22-rdf-syntax-ns#type"
TYPE_PATTERN = "http://rdf.basekb.com/ns/%s"


def main():
    writer = csv.writer(sys.stdout)

    seenPoliticians = set()

    conflicts = []
    not_found = []

    client = pyley.CayleyClient()
    g = pyley.GraphObject()

    for row in csv.reader(sys.stdin):
        politician_id = row[1]
        politician = row[2]

        if politician_id not in seenPoliticians:
            seenPoliticians.add(politician_id)

            query = g.V() \
                    .Has(NAME, ALIAS_PATTERN % politician) \
                    .Has(TYPE, TYPE_PATTERN % "government.politician") \
                    .All()

            response = client.Send(query)

            if not response.result.get('result'):
                # Try with alias
                query = g.V() \
                    .Has(ALIAS, ALIAS_PATTERN % politician) \
                    .Has(TYPE, TYPE_PATTERN % "government.politician") \
                    .All()

                response = client.Send(query)

            if not response.result.get('result'):
                # Empty response!
                #sys.stderr.write("No match found: %s\n" % politician)
                not_found.append({
                    'id': politician_id,
                    'name': politician,
                })
                continue

            res = response.result.get('result')

            if len(res) == 1:
                # Exact match B)
                p_id = res[0]['id']
                writer.writerow([politician_id, p_id])

            elif len(res) > 1:
                conflicts.append({
                    'id': politician_id,
                    'name': politician,
                    'choices': res,
                })
                continue

    # write conflicts into stderr.
    sys.stderr.write("%d conflicts\n" % len(conflicts))
    for conflict in conflicts:
        res = conflict['choices']
        politician = conflict['name']
        politician_id = conflict['id']
        sys.stderr.write("%s ; %s\n" % (politician_id, politician))
        i = 0
        for p in res:
            i += 1
            sys.stderr.write("\t%d. %s: \n" % (i, p['id']))

        sys.stderr.write("\n")

    sys.stderr.write("Not found: %d\n" % len(not_found))
    for r in not_found:
        sys.stderr.write("%s: %s\n" % (r['id'], r['name']))


if __name__ == "__main__":
    main()
