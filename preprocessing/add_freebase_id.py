# -*- coding: utf-8 -*-
import sys
import csv
import copy

reload(sys)
sys.setdefaultencoding('utf-8')

csv.field_size_limit(sys.maxsize)


def main():
    ids = {}
    with open('map_id_freebase_id.csv', 'rb') as f:
        for row in csv.reader(f):
            ids[row[0]] = row[1]

    writer = csv.writer(sys.stdout)

    for row in csv.reader(sys.stdin):
        if ids.get(row[1]):
            new_row = copy.copy(row)
            new_row.insert(2, ids[row[1]])
            writer.writerow(new_row)

if __name__ == "__main__":
    main()
