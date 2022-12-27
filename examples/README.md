## Examples

### CLI usage

Starting from a pseudo random `10000` bytes different compression algorithms produce different results

```
> cat /dev/random | dd bs=1 count=10000 2>/dev/null > 10000b-rand
```

```
> cat 10000b-rand | f2bist decode -s >/dev/null

bits: 80000

0: 40191
1: 39809


> cat 10000b-rand | f2bist decode -c b -s >/dev/null

bits: 80000

0: 40191
1: 39809

compression ratio: -0.040
compression algorithm: Brotli

bits: 80032

0: 40209
1: 39823

> cat 10000b-rand | f2bist decode -c s2 -s >/dev/null

bits: 80000

0: 40191
1: 39809

compression ratio: -0.180
compression algorithm: S2

bits: 80144

0: 40276
1: 39868
```

In random data, there is no structure that can be compressed to a simpler representation. On the other hand, compressing hamlet as `.txt` file takes is much more satisfying

```
> f2bist decode -s -c b hamlet.txt >/dev/null

bits: 1554696

0: 906313
1: 648383

compression ratio: 67.692
compression algorithm: Brotli

bits: 502296

0: 249800
1: 252496

> f2bist decode -s -c zstd hamlet.txt >/dev/null

bits: 1554344

0: 906115
1: 648229

compression ratio: 61.564
compression algorithm: Zstd

bits: 597424

0: 304010
1: 293414
```

To test the decoding/encoding don't alter the data I usually

```
> f2bist decode -str hamlet.txt | f2bist encode | diff hamlet.txt - ; echo $?
0
```

with compression

```
> f2bist decode -str -c b hamlet.txt | f2bist -c b encode >/dev/null | diff hamlet.txt - ; echo $?
0
```
