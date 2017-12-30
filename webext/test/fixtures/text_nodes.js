// It'd be nice to somehow automate the creation of the coordinates here

let base = {
  textContent: "\n    testing nodes",
  parentElement: {
    style: {}
  },
  bounding_box: {
    top: 10.1,
    bottom: 50.1,
    left: 10.1,
    right: 80.1
  },
  dom_rects: [{
    top: 10.1,
    left: 10.1,
    width: 70.1
  },
  {
    left: 10.1,
    top: 30.1,
    width: 50.1
  }]
};

export default [base];
