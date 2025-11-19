import { ref, resolveComponent, withCtx, createVNode, toDisplayString, createTextVNode, useSSRContext } from 'vue';
import { ssrRenderComponent, ssrInterpolate } from 'vue/server-renderer';
import { useRoute } from 'vue-router';

const _sfc_main = {
  __name: "[t]",
  __ssrInlineRender: true,
  setup(__props) {
    const route = useRoute();
    const title = ref({});
    const fetchTitle = async () => {
      const res = await fetch(`/api/titles/${route.params.t}`);
      title.value = await res.json();
    };
    fetchTitle();
    return (_ctx, _push, _parent, _attrs) => {
      const _component_UsaAccordion = resolveComponent("UsaAccordion");
      const _component_UsaAccordionItem = resolveComponent("UsaAccordionItem");
      _push(ssrRenderComponent(_component_UsaAccordion, _attrs, {
        default: withCtx((_, _push2, _parent2, _scopeId) => {
          if (_push2) {
            _push2(ssrRenderComponent(_component_UsaAccordionItem, { title: "Metrics" }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                if (_push3) {
                  _push3(`<p${_scopeId2}>Word Count: ${ssrInterpolate(title.value.total_words)}</p><p${_scopeId2}>RSCS: ${ssrInterpolate(title.value.avg_rscs)}</p>`);
                } else {
                  return [
                    createVNode("p", null, "Word Count: " + toDisplayString(title.value.total_words), 1),
                    createVNode("p", null, "RSCS: " + toDisplayString(title.value.avg_rscs), 1)
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
            _push2(ssrRenderComponent(_component_UsaAccordionItem, { title: "LSA Counts" }, null, _parent2, _scopeId));
            _push2(ssrRenderComponent(_component_UsaAccordionItem, { title: "Summary" }, {
              default: withCtx((_2, _push3, _parent3, _scopeId2) => {
                if (_push3) {
                  _push3(`${ssrInterpolate(title.value.summary)}`);
                } else {
                  return [
                    createTextVNode(toDisplayString(title.value.summary), 1)
                  ];
                }
              }),
              _: 1
            }, _parent2, _scopeId));
          } else {
            return [
              createVNode(_component_UsaAccordionItem, { title: "Metrics" }, {
                default: withCtx(() => [
                  createVNode("p", null, "Word Count: " + toDisplayString(title.value.total_words), 1),
                  createVNode("p", null, "RSCS: " + toDisplayString(title.value.avg_rscs), 1)
                ]),
                _: 1
              }),
              createVNode(_component_UsaAccordionItem, { title: "LSA Counts" }),
              createVNode(_component_UsaAccordionItem, { title: "Summary" }, {
                default: withCtx(() => [
                  createTextVNode(toDisplayString(title.value.summary), 1)
                ]),
                _: 1
              })
            ];
          }
        }),
        _: 1
      }, _parent));
    };
  }
};
const _sfc_setup = _sfc_main.setup;
_sfc_main.setup = (props, ctx) => {
  const ssrContext = useSSRContext();
  (ssrContext.modules || (ssrContext.modules = /* @__PURE__ */ new Set())).add("pages/title/[t].vue");
  return _sfc_setup ? _sfc_setup(props, ctx) : void 0;
};

export { _sfc_main as default };
//# sourceMappingURL=_t_-BWl8xKsz.mjs.map
