# -*- coding: utf-8 -*-
import re
import sys
import fileinput
import string
import nltk.corpus
import nltk.stem.snowball
from nltk import FreqDist

reload(sys)
sys.setdefaultencoding('utf-8')

STEMMER = nltk.stem.snowball.SnowballStemmer('english')
ENGLISH_STOPWORDS = set(nltk.corpus.stopwords.words('english'))
PUNCTUATION = string.punctuation + '“' + '”' + '—' + '…'
RE_START_WITH_PUNCT = re.compile(r'^(“|£|$|€|…|—|\.).+$')
RE_US_NUMBERS = re.compile(r'^[0-9]+,[0-9]+$')
VOC_NUMBER = 10000


def valid_word(word):
    test = True
    try:
        w = word
        if RE_US_NUMBERS.match(word):
            w = w.replace(',', '')
        float(w)
        return False
    except ValueError:
        pass
    test = (word[-1] not in PUNCTUATION
        and RE_START_WITH_PUNCT.match(word) is None)
    return word not in ENGLISH_STOPWORDS and word not in PUNCTUATION and test


def get_stems(sentence):
    words = sentence.split()
    for word in words:
        w = word.strip()
        if valid_word(w):
            yield STEMMER.stem(w)


def main():
    """
    Return the X most common stems from the dataset.
    X = VOC_NUMBER constant.
    """
    fdist = FreqDist()
    for line in fileinput.input():
        try:
            for stem in get_stems(line):
                if stem not in ENGLISH_STOPWORDS:
                    fdist.inc(stem)
        except UnicodeDecodeError:
            pass

    keys = fdist.keys()[:VOC_NUMBER]
    for s in keys:
        print s

if __name__ == "__main__":
    main()
