function a() {
  document.querySelectorAll("template[simple]").forEach((l) => {
    let s = l.getAttribute("id"), n = l.content;
    customElements.define(s, class extends HTMLElement {
      constructor() {
        super(), this.appendChild(n.cloneNode(!0)), this.slots = this.querySelectorAll("slot");
      }
      connectedCallback() {
        let o = [];
        this.slots.forEach((e) => {
          let r = e.getAttribute("name"), t = this.querySelector(`[slot="${r}"]`);
          t && (e.replaceWith(t.cloneNode(!0)), o.push(t));
        }), o.forEach((e) => {
          e.remove();
        });
      }
    });
  });
}
export {
  a as setup
};
