# -*- coding: utf-8 -*-
import sys
import re
import fileinput
import matplotlib.pyplot as plt

reload(sys)
sys.setdefaultencoding('utf-8')


def main():
    D = {}
    V = {}
    alias_re = re.compile(r'^[^"]+"([^"]+)"@en.+$')
    i = 0
    for line in fileinput.input():
        i += 1
        if i % 1000 == 0:
            sys.stderr.write("Line %d\n" % i)
        m = alias_re.match(line)
        if m:
            alias = m.group(1)
            if alias:
                D[alias] = D.get(alias, 0) + 1

    for value in D.itervalues():
        V[str(value)] = (value, V.get(str(value), (value, 0))[1] + 1)

    dist = []
    for value in V.itervalues():
        dist.append(value)

    fig = plt.figure()
    ax = fig.add_subplot(2, 1, 1)
    z = zip(*sorted(dist, key=lambda tup: tup[0]))
    print z
    ax.plot(*z)
    ax.set_yscale('log')
    plt.xlabel('|E(a)|')
    plt.show()

if __name__ == "__main__":
    main()
