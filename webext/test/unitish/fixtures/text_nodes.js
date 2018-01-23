// It'd be nice to somehow automate the creation of the coordinates here

let base = {
  textContent: "\n testing nodes",
  parentElement: {
    style: {}
  },
  bounding_box: {
    top: 0.1,
    bottom: 4.1,
    left: 0.1,
    right: 7.1,
    width: 7.1
  },
  dom_rects: [{
    // 'testing'
    top: 0.1,
    left: 0.1,
    width: 7.1
  },
  {
    // 'nodes'
    top: 4.1,
    left: 0.1,
    width: 5.1
  }]
};

export default [base];
