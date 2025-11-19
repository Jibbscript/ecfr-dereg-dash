import { ref, resolveComponent, withCtx, createTextVNode, toDisplayString, useSSRContext } from 'vue';
import { ssrRenderAttrs, ssrRenderComponent, ssrInterpolate } from 'vue/server-renderer';
import { useRoute } from 'vue-router';

const _sfc_main = {
  __name: "[id]",
  __ssrInlineRender: true,
  setup(__props) {
    const route = useRoute();
    const section = ref({});
    const fetchSection = async () => {
      const res = await fetch(`/api/sections/${route.params.id}`);
      section.value = await res.json();
    };
    fetchSection();
    return (_ctx, _push, _parent, _attrs) => {
      const _component_UsaHeading = resolveComponent("UsaHeading");
      _push(`<div${ssrRenderAttrs(_attrs)}>`);
      _push(ssrRenderComponent(_component_UsaHeading, null, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(`Section ${ssrInterpolate(section.value.section)}`);
          } else {
            return [
              createTextVNode("Section " + toDisplayString(section.value.section), 1)
            ];
          }
        }),
        _: 1
      }, _parent));
      _push(`<p>Excerpt: ${ssrInterpolate(section.value.text ? section.value.text.substring(0, 500) : "")}...</p><p>RSCS: ${ssrInterpolate(section.value.rscs_per_1k)}</p><p>Summary: ${ssrInterpolate(section.value.summary)}</p></div>`);
    };
  }
};
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/section/[id].vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=_id_-BpjKyz5A.mjs.map
