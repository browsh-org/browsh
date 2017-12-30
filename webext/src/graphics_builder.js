import BaseBuilder from 'base_builder';

// Converts an instance of the viewport into a scaled array of pixel values.
// Note, that it does this both with and without the text visible in order
// to aid in a clean seperation of the graphics and text in the final frame
// renderd in the terminal.
export default class GraphicsBuilder extends BaseBuilder {
  // Note that `frame_height` is twice the height of our TTY's height
  // because of the special UTF8 half-block trick that gets us 2 'pixels'
  // per character cell.
  constructor() {
    super();
    this.off_screen_canvas = document.createElement('canvas');
    this.ctx = this.off_screen_canvas.getContext('2d');
    this.updateCurrentViewportDimensions();
  }

  getSnapshotWithoutText(frame_width, frame_height) {
    this.frame_width = frame_width;
    this.frame_height = frame_height;
    this.hideText();
    let snapshot = this.getSnapshot();
    this.showText();
    return snapshot;
  }

  hideText() {
    this.styles = document.createElement("style");
    document.head.appendChild(this.styles);
    this.styles.sheet.insertRule(
      'html * { color: transparent !important; }'
    );
  }

  showText() {
    this.styles.parentNode.removeChild(this.styles);
  }


  getSnapshotWithText(frame_width, frame_height) {
    this.frame_width = frame_width;
    this.frame_height = frame_height;
    return this.getSnapshot();
  }

  getSnapshot() {
    this.updateCurrentViewportDimensions()
    this.scaleCanvas();
    let pixel_data = this.getPixelData();
    this.is_first_frame_finished = true;
    return pixel_data;
  }

  // Perhaps the window has been resized to better accomodate text-sizing, or to try
  // to trigger some mobile responsive CSS.
  // And of course also the page may have been scrolled.
  updateCurrentViewportDimensions() {
    this.viewport_width = window.innerWidth;
    this.viewport_height = window.innerHeight;
    this.page_x_position = window.scrollX;
    this.page_y_position = window.scrollY;
    // Resize our canvas to match the viewport. I guess this makes for efficient
    // use of memory?
    this.off_screen_canvas.width = this.viewport_width;
    this.off_screen_canvas.height = this.viewport_height;
  }

  // Scale the screenshot so that 1 pixel approximates one TTY cell.
  // TODO: Allow one of these to be manually adjusted in realtime through the Browsh client
  // in order for the user to set the correct aspect ratio for their particular terminal
  // setup.
  scaleCanvas() {
    let scale_x = this.frame_width / this.viewport_width;
    let scale_y = this.frame_height / this.viewport_height;
    this.ctx.scale(scale_x, scale_y);
  }

  // Get an array of RGB values.
  // This is Firefox-only. Chrome has a nicer MediaStream for this.
  getPixelData() {
    let background_colour = 'rgb(255,255,255)';
    this.ctx.drawWindow(
      window,
      this.page_x_position,
      this.page_y_position,
      this.viewport_width,
      this.viewport_height,
      background_colour
    );
    return this.ctx.getImageData(0, 0, this.frame_width, this.frame_height).data;
  }
}
