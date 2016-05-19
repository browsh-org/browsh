/*
   This code largely comes from Itai Nahshon's xzoom, see:
      http://git.r-36.net/xzoom
 */

#include <stdio.h>
#include <string.h>
#include <stdlib.h>
#include <unistd.h>

#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <X11/Xutil.h>

#include <X11/cursorfont.h>
#include <X11/keysym.h>


Display *dpy;
Screen  *scr;
Window   win;
Status   status;

Window root, child;
int    rootX, rootY, winX, winY;
unsigned int mask;

GC gc;

#define SRC  0      /* index for source image */
#define DST  1      /* index for dest image */

#define WIDTH  1600  /* default width */
#define HEIGHT 1200 /* default height */

#define MAG  1      /* default magnification */
#define MAGX MAG    /* horizontal magnification */
#define MAGY MAG    /* vertical magnification */

int xgrab, ygrab;   /* where do we take the picture from */

int magx = MAGX;
int magy = MAGY;

int old_magx = MAGX;

int gridx = False;
int gridy = False;

int width[2] = { 0, WIDTH };
int height[2] = { 0, HEIGHT };
unsigned depth = 0;

XImage *ximage[2]; /* Ximage struct. */

int created_images = False;

void allocate_images(void) {
  int i;

  for (i = 0; i < 2; i++) {
    char *data;
    data = malloc(BitmapUnit(dpy) / 8 * width[i] * height[i]);

    ximage[i] = XCreateImage(dpy,
                             DefaultVisualOfScreen(scr),
                             DefaultDepthOfScreen(scr),
                             ZPixmap, 0, data,
                             width[i], height[i], 32, 0);

    if (ximage[i] == NULL) {
      perror("XCreateImage");
      exit(-1);
    }
  }
  created_images = True;
}

void destroy_images(void) {
  int i;

  if (!created_images) return;

  for (i = 0; i < 2; i++) {
    free(ximage[i]->data);
    ximage[i]->data = NULL;   /* remove refrence to that address */
    XDestroyImage(ximage[i]); /* and destroy image */
  }

  created_images = False;
}

/* Resize is called with the dest size.
   Called when magnification changes or when
   actual window size is changed */
void resize(int new_width, int new_height) {
  destroy_images(); /* we can get rid of these */

  /* find new dimensions for source */
  width[SRC]  = (new_width + magx - 1) / magx;
  height[SRC] = (new_height + magy - 1) / magy;

  if (width[SRC] < 1) width[SRC] = 1;
  if (width[SRC] > WidthOfScreen(scr)) width[SRC] = WidthOfScreen(scr);
  if (height[SRC] < 1) height[SRC] = 1;
  if (height[SRC] > HeightOfScreen(scr)) height[SRC] = HeightOfScreen(scr);

  /* temporary, the dest image may be larger than the
     actual window */
  width[DST]  = magx * width[SRC];
  height[DST] = magy * height[SRC];

  allocate_images(); /* allocate new images */

  /* remember actual window size */
  if (width[DST] > new_width) width[DST] = new_width;
  if (height[DST] > new_height) height[DST] = new_height;

  // printf("x: %d, y: %d\n", rootX, rootY);
}

void scale8(void)
{
#define T unsigned char
#include "scale.h"
#undef T
}

void scale16(void)
{
#define T unsigned short
#include "scale.h"
#undef T
}

void scale32(void)
{
#define T unsigned int
#include "scale.h"
#undef T
}

void xzoom()
{
  XGetSubImage(dpy, RootWindowOfScreen(scr),
               xgrab, ygrab, width[SRC], height[SRC], AllPlanes,
               ZPixmap, ximage[SRC], 0, 0);

  if (depth == 8) scale8();
  else if (depth <= 8 * sizeof(short)) scale16();
  else if (depth <= 8 * sizeof(int)) scale32();

  XPutImage(dpy, win, gc, ximage[DST], 0, 0, 0, 0, width[DST], height[DST]);
  XSync(dpy, 0);
}

int xzoom_init() {
  XSetWindowAttributes xswa;
  XGCValues gcv;
  char *dpyname = ":0";
  int xpos = 0, ypos = 0;

  atexit(destroy_images);

  if (!(dpy = XOpenDisplay(dpyname))) {
    perror("Cannot open display");
    exit(-1);
  }

  scr = DefaultScreenOfDisplay(dpy);

  depth = DefaultDepthOfScreen(scr);

  // printf("=%dx%d+%d+%d\n", width[DST], height[DST], xpos, ypos);

  win = XCreateWindow(dpy, RootWindowOfScreen(scr),
                      WIDTH, 0, width[DST], height[DST], 0,
                      DefaultDepthOfScreen(scr), InputOutput,
                      DefaultVisualOfScreen(scr),
                      CWEventMask | CWBackPixel, &xswa);

  status = XMapWindow(dpy, win);

  gcv.plane_mask     = AllPlanes;
  gcv.subwindow_mode = IncludeInferiors;
  gcv.function       = GXcopy;
  gcv.foreground     = WhitePixelOfScreen(scr);
  gcv.background     = BlackPixelOfScreen(scr);
  gc                 = XCreateGC(dpy, RootWindowOfScreen(scr),
                                 GCFunction |
																 GCPlaneMask |
																 GCSubwindowMode |
																 GCForeground |
																 GCBackground,
                                 &gcv);

  resize(width[DST], height[DST]);
}

int xzoom_loop()
{

  if (old_magx != magx) {
    XQueryPointer(dpy, DefaultRootWindow(dpy), &root, &child,
      &mouseX, &mouseY, &winX, &winY, &mask);
    xgrab = mouseX - (width[SRC] / 2);
    ygrab = mouseY - (height[SRC] / 2);
    resize(width[DST], height[DST]);
  }
  old_magx = magx;

  if (xgrab < 0) xgrab = 0;

  if (xgrab > WidthOfScreen(scr) - width[SRC])
		xgrab = WidthOfScreen(scr) - width[SRC];

  if (ygrab < 0) ygrab = 0;

  if (ygrab > HeightOfScreen(scr) - height[SRC])
		ygrab = HeightOfScreen(scr) - height[SRC];

  xzoom();
}
