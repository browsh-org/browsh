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
Window   win, root, child;
Status   status;
int    winX, winY;
unsigned int mask;

GC gc;

#define SRC 0
#define DST 1

#define WIDTH  1600
#define HEIGHT 1200

// The top left of the area that we zoom in on
int xgrab = 0;
int ygrab = 0;

int magnification = 1;
int old_magnification = 1;

int width[2] = { WIDTH, WIDTH };
int height[2] = { HEIGHT, HEIGHT };
unsigned depth = 0;

XImage *ximage[2];

int created_images = False;

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
    ximage[i]->data = NULL;
    XDestroyImage(ximage[i]);
  }

  created_images = False;
}

// Update the zoom window with the current state of the desktop.
// Happens at 25fps.
void update_zoom_window_with_desktop() {
  // Get a snapshot of the desktop, or a portion of the desktop
  XGetSubImage(dpy, RootWindowOfScreen(scr),
               xgrab, ygrab, width[SRC], height[SRC], AllPlanes,
               ZPixmap, ximage[SRC], 0, 0);

  // Zoom in on that snapshot
  if (depth <= 8 * sizeof(short)) scale16();
  else if (depth <= 8 * sizeof(int)) scale32();

  // Put the snapshot into the xzoom window
  XPutImage(dpy, win, gc, ximage[DST], 0, 0, 0, 0, width[DST], height[DST]);
  XSync(dpy, 0);
}

void recreate_images_on_zoom() {
  if (old_magnification != magnification) {
    destroy_images();
    allocate_images();
  }
  old_magnification = magnification;
}

void loop() {
  recreate_images_on_zoom();
  update_zoom_window_with_desktop();
}

int xzoom_init() {
  XSetWindowAttributes xswa;
  XGCValues gcv;
  char *dpyname = ":0";

  atexit(destroy_images);

  if (!(dpy = XOpenDisplay(dpyname))) {
    perror("Cannot open display");
    exit(-1);
  }

  scr = DefaultScreenOfDisplay(dpy);

  depth = DefaultDepthOfScreen(scr);

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
  allocate_images();
}
