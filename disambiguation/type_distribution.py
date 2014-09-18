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
    type_re = re.compile(r'^.+<.+\/ns\/([^>]+)>.+$')
    for line in fileinput.input():
        m = type_re.match(line)
        if m:
            t = m.group(1)
            if t:
                D[t] = D.get(t, 0) + 1

    for value in D.itervalues():
        V[str(value)] = (value, V.get(str(value), (value, 0))[1] + 1)

    dist = []
    for value in V.itervalues():
        dist.append(value)

    fig = plt.figure()
    ax = fig.add_subplot(2, 1, 1)
    z = zip(*sorted(dist, key=lambda tup: tup[0]))
    ax.plot(*z)
    ax.set_yscale('log')
    ax.set_xscale('log')
    plt.xlabel('|T(e)|')
    plt.show()

if __name__ == "__main__":
    main()
