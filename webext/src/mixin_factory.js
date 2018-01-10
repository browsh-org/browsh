export default function (...mixins) {
  return mixins.reduce((base, mixin) => {
    return mixin(base);
  }, class {});
}
