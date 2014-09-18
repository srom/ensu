# -*- coding: utf-8 -*-
import sys
import csv
import nltk.data

reload(sys)
sys.setdefaultencoding('utf-8')


def main():
    writer = csv.writer(sys.stdout)

    sent_detector = nltk.data.load('tokenizers/punkt/english.pickle')

    count = 0
    for row in csv.reader(sys.stdin):
        count += 1
        # if count < 58400:
        #     continue
        if count % 100 == 0:
            sys.stderr.write("%d\n" % count)
        try:
            doc_id = row[0]  # document id
            text = row[5].replace("\\n", "\n")
            text = text.replace(" hon. ", " honourable ")
            sn = -1
            for sentence in sent_detector.tokenize(text.strip()):
                sn += 1
                sent_id = doc_id + '__%d' % sn
                writer.writerow([
                    sent_id, row[0], row[1], row[2], row[3], row[4], sentence])

        except UnicodeDecodeError:
            pass

if __name__ == "__main__":
    main()
