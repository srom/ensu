# -*- coding: utf-8 -*-
import sys
import csv
import copy
import string
import nltk.data
import nltk.corpus
import nltk.tokenize
from nltk.util import ngrams

reload(sys)
sys.setdefaultencoding('utf-8')

ENGLISH_STOPWORDS = set(nltk.corpus.stopwords.words('english'))


def create_candidates_lists(tokens):
    """
    Split the tokens in lists everytime we encounter puncutation token.
    e.g: ['hello', ',', 'hey there', '!'] --> [['hello'],['hey', 'there']]
    """
    res = []
    l = []
    for token in tokens:
        if token in string.punctuation and len(l):
            res.append(copy.copy(l))
            l = []
        elif token not in string.punctuation:
            l.append(token)
    return res


def uppercase_first_letters(ngram):
    """
    Uppercase first letters of words in a given ngram, except for stopwords.
    (except if stopword is the first word)
    e.g: "welfare reform in europe" --> "Welfare Reform in Europe"
         "you are welcome" --> "You are Welcome"
    """
    res = []
    idx = 0
    long_word = True if len(ngram.split()) > 1 else False
    for w in ngram.split():
        if long_word and w in ENGLISH_STOPWORDS and idx != 0:
            res.append(w)
        else:
            res.append(w[0].upper() + w[1:])
        idx += 1
    return ' '.join(res)


def create_candidate_list(sentence):
    tokens = nltk.tokenize.word_tokenize(sentence)

    candidates_lists = create_candidates_lists(tokens)

    # Create list of 1-grams.
    candidates = []
    for l in candidates_lists:
        candidates += l

    # Remove irrelevant stop words in 1-grams.
    res = [token for token in candidates
        if token not in ENGLISH_STOPWORDS]

    # Create list of bigrams.
    bigrams = []
    for l in candidates_lists:
        bigrams += ngrams(l, 2)

    # Create list of trigrams.
    trigrams = []
    for l in candidates_lists:
        trigrams += ngrams(l, 3)

    # Create list of 4-grams.
    fourgrams = []
    for l in candidates_lists:
        fourgrams += ngrams(l, 4)

    res += [' '.join(a) for a in bigrams]
    res += [' '.join(a) for a in trigrams]
    res += [' '.join(a) for a in fourgrams]

    return res


def main():
    sent_detector = nltk.data.load('tokenizers/punkt/english.pickle')
    aliases = set()
    for row in csv.reader(sys.stdin):
        try:
            text = row[4].replace("\\n", "\n")
            text = text.replace(" hon. ", " honourable ")  # Quick fix
            for sentence in sent_detector.tokenize(text.strip()):
                res = create_candidate_list(sentence)

                # Output one gram per line.
                for el in res:
                    a = uppercase_first_letters(el)
                    if a not in aliases:
                        aliases.add(a)
                        print a
        except UnicodeDecodeError:
            pass

if __name__ == "__main__":
    main()
