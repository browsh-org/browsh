/* scale image from SRC to DST - parameterized by type T */

/* get pixel address of point (x,y) in image t */
#define getP(t, x, y)                                             \
  (T *)(&ximage[t]->data[(ximage[t]->xoffset + (x)) * sizeof(T) + \
                         (y) * ximage[t]->bytes_per_line])

{
  int i, j, k;

  /* copy scaled lines from SRC to DST */
  j = height[SRC] - 1;

  do {
    T  *p1;
    T  *p2;
    int p2step;
    T  *p1_save;

    /* p1 point to beginning of scanline j*magnification in DST */
    p1      = getP(DST, 0, j * magnification);
    p1_save = p1;

    /* p2 point to beginning of scanline j in SRC */
    p2 = getP(SRC, 0, j);

    i = width[SRC];

    do {
      T c = *p2++;
      k = magnification;

      do *p1++ = c; while (--k > 0);
    } while (--i > 0);

    /* duplicate that line as needed */
    if (magnification > 1)
    {
      /* p1 point to beginning of scanline j*magnification in DST */
      p1 = p1_save;

      /* p2 points to beginning of next line */
      p2     = p1;
      p2step = ximage[DST]->bytes_per_line / sizeof(T);

      i = width[DST] * sizeof(T);
      k = magnification - 1;

      do {
        p2 += p2step;
        memcpy(p2, p1, i);
      } while (--k > 0);
    }
  } while (--j >= 0);
}

#undef getP
