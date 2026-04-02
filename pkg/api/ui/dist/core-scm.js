/**
 * @license
 * Copyright 2019 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const K = globalThis, re = K.ShadowRoot && (K.ShadyCSS === void 0 || K.ShadyCSS.nativeShadow) && "adoptedStyleSheets" in Document.prototype && "replace" in CSSStyleSheet.prototype, ie = Symbol(), ne = /* @__PURE__ */ new WeakMap();
let ve = class {
  constructor(e, s, i) {
    if (this._$cssResult$ = !0, i !== ie) throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");
    this.cssText = e, this.t = s;
  }
  get styleSheet() {
    let e = this.o;
    const s = this.t;
    if (re && e === void 0) {
      const i = s !== void 0 && s.length === 1;
      i && (e = ne.get(s)), e === void 0 && ((this.o = e = new CSSStyleSheet()).replaceSync(this.cssText), i && ne.set(s, e));
    }
    return e;
  }
  toString() {
    return this.cssText;
  }
};
const Ae = (t) => new ve(typeof t == "string" ? t : t + "", void 0, ie), B = (t, ...e) => {
  const s = t.length === 1 ? t[0] : e.reduce((i, r, o) => i + ((a) => {
    if (a._$cssResult$ === !0) return a.cssText;
    if (typeof a == "number") return a;
    throw Error("Value passed to 'css' function must be a 'css' function result: " + a + ". Use 'unsafeCSS' to pass non-literal values, but take care to ensure page security.");
  })(r) + t[o + 1], t[0]);
  return new ve(s, t, ie);
}, Se = (t, e) => {
  if (re) t.adoptedStyleSheets = e.map((s) => s instanceof CSSStyleSheet ? s : s.styleSheet);
  else for (const s of e) {
    const i = document.createElement("style"), r = K.litNonce;
    r !== void 0 && i.setAttribute("nonce", r), i.textContent = s.cssText, t.appendChild(i);
  }
}, le = re ? (t) => t : (t) => t instanceof CSSStyleSheet ? ((e) => {
  let s = "";
  for (const i of e.cssRules) s += i.cssText;
  return Ae(s);
})(t) : t;
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const { is: Ce, defineProperty: Ee, getOwnPropertyDescriptor: Ue, getOwnPropertyNames: Pe, getOwnPropertySymbols: Oe, getPrototypeOf: ze } = Object, k = globalThis, de = k.trustedTypes, Re = de ? de.emptyScript : "", X = k.reactiveElementPolyfillSupport, H = (t, e) => t, J = { toAttribute(t, e) {
  switch (e) {
    case Boolean:
      t = t ? Re : null;
      break;
    case Object:
    case Array:
      t = t == null ? t : JSON.stringify(t);
  }
  return t;
}, fromAttribute(t, e) {
  let s = t;
  switch (e) {
    case Boolean:
      s = t !== null;
      break;
    case Number:
      s = t === null ? null : Number(t);
      break;
    case Object:
    case Array:
      try {
        s = JSON.parse(t);
      } catch {
        s = null;
      }
  }
  return s;
} }, ae = (t, e) => !Ce(t, e), ce = { attribute: !0, type: String, converter: J, reflect: !1, useDefault: !1, hasChanged: ae };
Symbol.metadata ?? (Symbol.metadata = Symbol("metadata")), k.litPropertyMetadata ?? (k.litPropertyMetadata = /* @__PURE__ */ new WeakMap());
let R = class extends HTMLElement {
  static addInitializer(e) {
    this._$Ei(), (this.l ?? (this.l = [])).push(e);
  }
  static get observedAttributes() {
    return this.finalize(), this._$Eh && [...this._$Eh.keys()];
  }
  static createProperty(e, s = ce) {
    if (s.state && (s.attribute = !1), this._$Ei(), this.prototype.hasOwnProperty(e) && ((s = Object.create(s)).wrapped = !0), this.elementProperties.set(e, s), !s.noAccessor) {
      const i = Symbol(), r = this.getPropertyDescriptor(e, i, s);
      r !== void 0 && Ee(this.prototype, e, r);
    }
  }
  static getPropertyDescriptor(e, s, i) {
    const { get: r, set: o } = Ue(this.prototype, e) ?? { get() {
      return this[s];
    }, set(a) {
      this[s] = a;
    } };
    return { get: r, set(a) {
      const d = r == null ? void 0 : r.call(this);
      o == null || o.call(this, a), this.requestUpdate(e, d, i);
    }, configurable: !0, enumerable: !0 };
  }
  static getPropertyOptions(e) {
    return this.elementProperties.get(e) ?? ce;
  }
  static _$Ei() {
    if (this.hasOwnProperty(H("elementProperties"))) return;
    const e = ze(this);
    e.finalize(), e.l !== void 0 && (this.l = [...e.l]), this.elementProperties = new Map(e.elementProperties);
  }
  static finalize() {
    if (this.hasOwnProperty(H("finalized"))) return;
    if (this.finalized = !0, this._$Ei(), this.hasOwnProperty(H("properties"))) {
      const s = this.properties, i = [...Pe(s), ...Oe(s)];
      for (const r of i) this.createProperty(r, s[r]);
    }
    const e = this[Symbol.metadata];
    if (e !== null) {
      const s = litPropertyMetadata.get(e);
      if (s !== void 0) for (const [i, r] of s) this.elementProperties.set(i, r);
    }
    this._$Eh = /* @__PURE__ */ new Map();
    for (const [s, i] of this.elementProperties) {
      const r = this._$Eu(s, i);
      r !== void 0 && this._$Eh.set(r, s);
    }
    this.elementStyles = this.finalizeStyles(this.styles);
  }
  static finalizeStyles(e) {
    const s = [];
    if (Array.isArray(e)) {
      const i = new Set(e.flat(1 / 0).reverse());
      for (const r of i) s.unshift(le(r));
    } else e !== void 0 && s.push(le(e));
    return s;
  }
  static _$Eu(e, s) {
    const i = s.attribute;
    return i === !1 ? void 0 : typeof i == "string" ? i : typeof e == "string" ? e.toLowerCase() : void 0;
  }
  constructor() {
    super(), this._$Ep = void 0, this.isUpdatePending = !1, this.hasUpdated = !1, this._$Em = null, this._$Ev();
  }
  _$Ev() {
    var e;
    this._$ES = new Promise((s) => this.enableUpdating = s), this._$AL = /* @__PURE__ */ new Map(), this._$E_(), this.requestUpdate(), (e = this.constructor.l) == null || e.forEach((s) => s(this));
  }
  addController(e) {
    var s;
    (this._$EO ?? (this._$EO = /* @__PURE__ */ new Set())).add(e), this.renderRoot !== void 0 && this.isConnected && ((s = e.hostConnected) == null || s.call(e));
  }
  removeController(e) {
    var s;
    (s = this._$EO) == null || s.delete(e);
  }
  _$E_() {
    const e = /* @__PURE__ */ new Map(), s = this.constructor.elementProperties;
    for (const i of s.keys()) this.hasOwnProperty(i) && (e.set(i, this[i]), delete this[i]);
    e.size > 0 && (this._$Ep = e);
  }
  createRenderRoot() {
    const e = this.shadowRoot ?? this.attachShadow(this.constructor.shadowRootOptions);
    return Se(e, this.constructor.elementStyles), e;
  }
  connectedCallback() {
    var e;
    this.renderRoot ?? (this.renderRoot = this.createRenderRoot()), this.enableUpdating(!0), (e = this._$EO) == null || e.forEach((s) => {
      var i;
      return (i = s.hostConnected) == null ? void 0 : i.call(s);
    });
  }
  enableUpdating(e) {
  }
  disconnectedCallback() {
    var e;
    (e = this._$EO) == null || e.forEach((s) => {
      var i;
      return (i = s.hostDisconnected) == null ? void 0 : i.call(s);
    });
  }
  attributeChangedCallback(e, s, i) {
    this._$AK(e, i);
  }
  _$ET(e, s) {
    var o;
    const i = this.constructor.elementProperties.get(e), r = this.constructor._$Eu(e, i);
    if (r !== void 0 && i.reflect === !0) {
      const a = (((o = i.converter) == null ? void 0 : o.toAttribute) !== void 0 ? i.converter : J).toAttribute(s, i.type);
      this._$Em = e, a == null ? this.removeAttribute(r) : this.setAttribute(r, a), this._$Em = null;
    }
  }
  _$AK(e, s) {
    var o, a;
    const i = this.constructor, r = i._$Eh.get(e);
    if (r !== void 0 && this._$Em !== r) {
      const d = i.getPropertyOptions(r), n = typeof d.converter == "function" ? { fromAttribute: d.converter } : ((o = d.converter) == null ? void 0 : o.fromAttribute) !== void 0 ? d.converter : J;
      this._$Em = r;
      const h = n.fromAttribute(s, d.type);
      this[r] = h ?? ((a = this._$Ej) == null ? void 0 : a.get(r)) ?? h, this._$Em = null;
    }
  }
  requestUpdate(e, s, i, r = !1, o) {
    var a;
    if (e !== void 0) {
      const d = this.constructor;
      if (r === !1 && (o = this[e]), i ?? (i = d.getPropertyOptions(e)), !((i.hasChanged ?? ae)(o, s) || i.useDefault && i.reflect && o === ((a = this._$Ej) == null ? void 0 : a.get(e)) && !this.hasAttribute(d._$Eu(e, i)))) return;
      this.C(e, s, i);
    }
    this.isUpdatePending === !1 && (this._$ES = this._$EP());
  }
  C(e, s, { useDefault: i, reflect: r, wrapped: o }, a) {
    i && !(this._$Ej ?? (this._$Ej = /* @__PURE__ */ new Map())).has(e) && (this._$Ej.set(e, a ?? s ?? this[e]), o !== !0 || a !== void 0) || (this._$AL.has(e) || (this.hasUpdated || i || (s = void 0), this._$AL.set(e, s)), r === !0 && this._$Em !== e && (this._$Eq ?? (this._$Eq = /* @__PURE__ */ new Set())).add(e));
  }
  async _$EP() {
    this.isUpdatePending = !0;
    try {
      await this._$ES;
    } catch (s) {
      Promise.reject(s);
    }
    const e = this.scheduleUpdate();
    return e != null && await e, !this.isUpdatePending;
  }
  scheduleUpdate() {
    return this.performUpdate();
  }
  performUpdate() {
    var i;
    if (!this.isUpdatePending) return;
    if (!this.hasUpdated) {
      if (this.renderRoot ?? (this.renderRoot = this.createRenderRoot()), this._$Ep) {
        for (const [o, a] of this._$Ep) this[o] = a;
        this._$Ep = void 0;
      }
      const r = this.constructor.elementProperties;
      if (r.size > 0) for (const [o, a] of r) {
        const { wrapped: d } = a, n = this[o];
        d !== !0 || this._$AL.has(o) || n === void 0 || this.C(o, void 0, a, n);
      }
    }
    let e = !1;
    const s = this._$AL;
    try {
      e = this.shouldUpdate(s), e ? (this.willUpdate(s), (i = this._$EO) == null || i.forEach((r) => {
        var o;
        return (o = r.hostUpdate) == null ? void 0 : o.call(r);
      }), this.update(s)) : this._$EM();
    } catch (r) {
      throw e = !1, this._$EM(), r;
    }
    e && this._$AE(s);
  }
  willUpdate(e) {
  }
  _$AE(e) {
    var s;
    (s = this._$EO) == null || s.forEach((i) => {
      var r;
      return (r = i.hostUpdated) == null ? void 0 : r.call(i);
    }), this.hasUpdated || (this.hasUpdated = !0, this.firstUpdated(e)), this.updated(e);
  }
  _$EM() {
    this._$AL = /* @__PURE__ */ new Map(), this.isUpdatePending = !1;
  }
  get updateComplete() {
    return this.getUpdateComplete();
  }
  getUpdateComplete() {
    return this._$ES;
  }
  shouldUpdate(e) {
    return !0;
  }
  update(e) {
    this._$Eq && (this._$Eq = this._$Eq.forEach((s) => this._$ET(s, this[s]))), this._$EM();
  }
  updated(e) {
  }
  firstUpdated(e) {
  }
};
R.elementStyles = [], R.shadowRootOptions = { mode: "open" }, R[H("elementProperties")] = /* @__PURE__ */ new Map(), R[H("finalized")] = /* @__PURE__ */ new Map(), X == null || X({ ReactiveElement: R }), (k.reactiveElementVersions ?? (k.reactiveElementVersions = [])).push("2.1.2");
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const D = globalThis, pe = (t) => t, Q = D.trustedTypes, he = Q ? Q.createPolicy("lit-html", { createHTML: (t) => t }) : void 0, $e = "$lit$", _ = `lit$${Math.random().toFixed(9).slice(2)}$`, xe = "?" + _, Me = `<${xe}>`, O = document, L = () => O.createComment(""), q = (t) => t === null || typeof t != "object" && typeof t != "function", oe = Array.isArray, Te = (t) => oe(t) || typeof (t == null ? void 0 : t[Symbol.iterator]) == "function", ee = `[
\f\r]`, j = /<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g, me = /-->/g, ue = />/g, E = RegExp(`>|${ee}(?:([^\\s"'>=/]+)(${ee}*=${ee}*(?:[^
\f\r"'\`<>=]|("|')|))|$)`, "g"), fe = /'/g, ge = /"/g, we = /^(?:script|style|textarea|title)$/i, Ie = (t) => (e, ...s) => ({ _$litType$: t, strings: e, values: s }), l = Ie(1), M = Symbol.for("lit-noChange"), c = Symbol.for("lit-nothing"), be = /* @__PURE__ */ new WeakMap(), U = O.createTreeWalker(O, 129);
function _e(t, e) {
  if (!oe(t) || !t.hasOwnProperty("raw")) throw Error("invalid template strings array");
  return he !== void 0 ? he.createHTML(e) : e;
}
const Ne = (t, e) => {
  const s = t.length - 1, i = [];
  let r, o = e === 2 ? "<svg>" : e === 3 ? "<math>" : "", a = j;
  for (let d = 0; d < s; d++) {
    const n = t[d];
    let h, u, p = -1, f = 0;
    for (; f < n.length && (a.lastIndex = f, u = a.exec(n), u !== null); ) f = a.lastIndex, a === j ? u[1] === "!--" ? a = me : u[1] !== void 0 ? a = ue : u[2] !== void 0 ? (we.test(u[2]) && (r = RegExp("</" + u[2], "g")), a = E) : u[3] !== void 0 && (a = E) : a === E ? u[0] === ">" ? (a = r ?? j, p = -1) : u[1] === void 0 ? p = -2 : (p = a.lastIndex - u[2].length, h = u[1], a = u[3] === void 0 ? E : u[3] === '"' ? ge : fe) : a === ge || a === fe ? a = E : a === me || a === ue ? a = j : (a = E, r = void 0);
    const b = a === E && t[d + 1].startsWith("/>") ? " " : "";
    o += a === j ? n + Me : p >= 0 ? (i.push(h), n.slice(0, p) + $e + n.slice(p) + _ + b) : n + _ + (p === -2 ? d : b);
  }
  return [_e(t, o + (t[s] || "<?>") + (e === 2 ? "</svg>" : e === 3 ? "</math>" : "")), i];
};
class W {
  constructor({ strings: e, _$litType$: s }, i) {
    let r;
    this.parts = [];
    let o = 0, a = 0;
    const d = e.length - 1, n = this.parts, [h, u] = Ne(e, s);
    if (this.el = W.createElement(h, i), U.currentNode = this.el.content, s === 2 || s === 3) {
      const p = this.el.content.firstChild;
      p.replaceWith(...p.childNodes);
    }
    for (; (r = U.nextNode()) !== null && n.length < d; ) {
      if (r.nodeType === 1) {
        if (r.hasAttributes()) for (const p of r.getAttributeNames()) if (p.endsWith($e)) {
          const f = u[a++], b = r.getAttribute(p).split(_), y = /([.?@])?(.*)/.exec(f);
          n.push({ type: 1, index: o, name: y[2], strings: b, ctor: y[1] === "." ? He : y[1] === "?" ? De : y[1] === "@" ? Le : Z }), r.removeAttribute(p);
        } else p.startsWith(_) && (n.push({ type: 6, index: o }), r.removeAttribute(p));
        if (we.test(r.tagName)) {
          const p = r.textContent.split(_), f = p.length - 1;
          if (f > 0) {
            r.textContent = Q ? Q.emptyScript : "";
            for (let b = 0; b < f; b++) r.append(p[b], L()), U.nextNode(), n.push({ type: 2, index: ++o });
            r.append(p[f], L());
          }
        }
      } else if (r.nodeType === 8) if (r.data === xe) n.push({ type: 2, index: o });
      else {
        let p = -1;
        for (; (p = r.data.indexOf(_, p + 1)) !== -1; ) n.push({ type: 7, index: o }), p += _.length - 1;
      }
      o++;
    }
  }
  static createElement(e, s) {
    const i = O.createElement("template");
    return i.innerHTML = e, i;
  }
}
function T(t, e, s = t, i) {
  var a, d;
  if (e === M) return e;
  let r = i !== void 0 ? (a = s._$Co) == null ? void 0 : a[i] : s._$Cl;
  const o = q(e) ? void 0 : e._$litDirective$;
  return (r == null ? void 0 : r.constructor) !== o && ((d = r == null ? void 0 : r._$AO) == null || d.call(r, !1), o === void 0 ? r = void 0 : (r = new o(t), r._$AT(t, s, i)), i !== void 0 ? (s._$Co ?? (s._$Co = []))[i] = r : s._$Cl = r), r !== void 0 && (e = T(t, r._$AS(t, e.values), r, i)), e;
}
class je {
  constructor(e, s) {
    this._$AV = [], this._$AN = void 0, this._$AD = e, this._$AM = s;
  }
  get parentNode() {
    return this._$AM.parentNode;
  }
  get _$AU() {
    return this._$AM._$AU;
  }
  u(e) {
    const { el: { content: s }, parts: i } = this._$AD, r = ((e == null ? void 0 : e.creationScope) ?? O).importNode(s, !0);
    U.currentNode = r;
    let o = U.nextNode(), a = 0, d = 0, n = i[0];
    for (; n !== void 0; ) {
      if (a === n.index) {
        let h;
        n.type === 2 ? h = new F(o, o.nextSibling, this, e) : n.type === 1 ? h = new n.ctor(o, n.name, n.strings, this, e) : n.type === 6 && (h = new qe(o, this, e)), this._$AV.push(h), n = i[++d];
      }
      a !== (n == null ? void 0 : n.index) && (o = U.nextNode(), a++);
    }
    return U.currentNode = O, r;
  }
  p(e) {
    let s = 0;
    for (const i of this._$AV) i !== void 0 && (i.strings !== void 0 ? (i._$AI(e, i, s), s += i.strings.length - 2) : i._$AI(e[s])), s++;
  }
}
class F {
  get _$AU() {
    var e;
    return ((e = this._$AM) == null ? void 0 : e._$AU) ?? this._$Cv;
  }
  constructor(e, s, i, r) {
    this.type = 2, this._$AH = c, this._$AN = void 0, this._$AA = e, this._$AB = s, this._$AM = i, this.options = r, this._$Cv = (r == null ? void 0 : r.isConnected) ?? !0;
  }
  get parentNode() {
    let e = this._$AA.parentNode;
    const s = this._$AM;
    return s !== void 0 && (e == null ? void 0 : e.nodeType) === 11 && (e = s.parentNode), e;
  }
  get startNode() {
    return this._$AA;
  }
  get endNode() {
    return this._$AB;
  }
  _$AI(e, s = this) {
    e = T(this, e, s), q(e) ? e === c || e == null || e === "" ? (this._$AH !== c && this._$AR(), this._$AH = c) : e !== this._$AH && e !== M && this._(e) : e._$litType$ !== void 0 ? this.$(e) : e.nodeType !== void 0 ? this.T(e) : Te(e) ? this.k(e) : this._(e);
  }
  O(e) {
    return this._$AA.parentNode.insertBefore(e, this._$AB);
  }
  T(e) {
    this._$AH !== e && (this._$AR(), this._$AH = this.O(e));
  }
  _(e) {
    this._$AH !== c && q(this._$AH) ? this._$AA.nextSibling.data = e : this.T(O.createTextNode(e)), this._$AH = e;
  }
  $(e) {
    var o;
    const { values: s, _$litType$: i } = e, r = typeof i == "number" ? this._$AC(e) : (i.el === void 0 && (i.el = W.createElement(_e(i.h, i.h[0]), this.options)), i);
    if (((o = this._$AH) == null ? void 0 : o._$AD) === r) this._$AH.p(s);
    else {
      const a = new je(r, this), d = a.u(this.options);
      a.p(s), this.T(d), this._$AH = a;
    }
  }
  _$AC(e) {
    let s = be.get(e.strings);
    return s === void 0 && be.set(e.strings, s = new W(e)), s;
  }
  k(e) {
    oe(this._$AH) || (this._$AH = [], this._$AR());
    const s = this._$AH;
    let i, r = 0;
    for (const o of e) r === s.length ? s.push(i = new F(this.O(L()), this.O(L()), this, this.options)) : i = s[r], i._$AI(o), r++;
    r < s.length && (this._$AR(i && i._$AB.nextSibling, r), s.length = r);
  }
  _$AR(e = this._$AA.nextSibling, s) {
    var i;
    for ((i = this._$AP) == null ? void 0 : i.call(this, !1, !0, s); e !== this._$AB; ) {
      const r = pe(e).nextSibling;
      pe(e).remove(), e = r;
    }
  }
  setConnected(e) {
    var s;
    this._$AM === void 0 && (this._$Cv = e, (s = this._$AP) == null || s.call(this, e));
  }
}
class Z {
  get tagName() {
    return this.element.tagName;
  }
  get _$AU() {
    return this._$AM._$AU;
  }
  constructor(e, s, i, r, o) {
    this.type = 1, this._$AH = c, this._$AN = void 0, this.element = e, this.name = s, this._$AM = r, this.options = o, i.length > 2 || i[0] !== "" || i[1] !== "" ? (this._$AH = Array(i.length - 1).fill(new String()), this.strings = i) : this._$AH = c;
  }
  _$AI(e, s = this, i, r) {
    const o = this.strings;
    let a = !1;
    if (o === void 0) e = T(this, e, s, 0), a = !q(e) || e !== this._$AH && e !== M, a && (this._$AH = e);
    else {
      const d = e;
      let n, h;
      for (e = o[0], n = 0; n < o.length - 1; n++) h = T(this, d[i + n], s, n), h === M && (h = this._$AH[n]), a || (a = !q(h) || h !== this._$AH[n]), h === c ? e = c : e !== c && (e += (h ?? "") + o[n + 1]), this._$AH[n] = h;
    }
    a && !r && this.j(e);
  }
  j(e) {
    e === c ? this.element.removeAttribute(this.name) : this.element.setAttribute(this.name, e ?? "");
  }
}
class He extends Z {
  constructor() {
    super(...arguments), this.type = 3;
  }
  j(e) {
    this.element[this.name] = e === c ? void 0 : e;
  }
}
class De extends Z {
  constructor() {
    super(...arguments), this.type = 4;
  }
  j(e) {
    this.element.toggleAttribute(this.name, !!e && e !== c);
  }
}
class Le extends Z {
  constructor(e, s, i, r, o) {
    super(e, s, i, r, o), this.type = 5;
  }
  _$AI(e, s = this) {
    if ((e = T(this, e, s, 0) ?? c) === M) return;
    const i = this._$AH, r = e === c && i !== c || e.capture !== i.capture || e.once !== i.once || e.passive !== i.passive, o = e !== c && (i === c || r);
    r && this.element.removeEventListener(this.name, this, i), o && this.element.addEventListener(this.name, this, e), this._$AH = e;
  }
  handleEvent(e) {
    var s;
    typeof this._$AH == "function" ? this._$AH.call(((s = this.options) == null ? void 0 : s.host) ?? this.element, e) : this._$AH.handleEvent(e);
  }
}
class qe {
  constructor(e, s, i) {
    this.element = e, this.type = 6, this._$AN = void 0, this._$AM = s, this.options = i;
  }
  get _$AU() {
    return this._$AM._$AU;
  }
  _$AI(e) {
    T(this, e);
  }
}
const te = D.litHtmlPolyfillSupport;
te == null || te(W, F), (D.litHtmlVersions ?? (D.litHtmlVersions = [])).push("3.3.2");
const We = (t, e, s) => {
  const i = (s == null ? void 0 : s.renderBefore) ?? e;
  let r = i._$litPart$;
  if (r === void 0) {
    const o = (s == null ? void 0 : s.renderBefore) ?? null;
    i._$litPart$ = r = new F(e.insertBefore(L(), o), o, void 0, s ?? {});
  }
  return r._$AI(t), r;
};
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const P = globalThis;
class x extends R {
  constructor() {
    super(...arguments), this.renderOptions = { host: this }, this._$Do = void 0;
  }
  createRenderRoot() {
    var s;
    const e = super.createRenderRoot();
    return (s = this.renderOptions).renderBefore ?? (s.renderBefore = e.firstChild), e;
  }
  update(e) {
    const s = this.render();
    this.hasUpdated || (this.renderOptions.isConnected = this.isConnected), super.update(e), this._$Do = We(s, this.renderRoot, this.renderOptions);
  }
  connectedCallback() {
    var e;
    super.connectedCallback(), (e = this._$Do) == null || e.setConnected(!0);
  }
  disconnectedCallback() {
    var e;
    super.disconnectedCallback(), (e = this._$Do) == null || e.setConnected(!1);
  }
  render() {
    return M;
  }
}
var ye;
x._$litElement$ = !0, x.finalized = !0, (ye = P.litElementHydrateSupport) == null || ye.call(P, { LitElement: x });
const se = P.litElementPolyfillSupport;
se == null || se({ LitElement: x });
(P.litElementVersions ?? (P.litElementVersions = [])).push("4.2.2");
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const V = (t) => (e, s) => {
  s !== void 0 ? s.addInitializer(() => {
    customElements.define(t, e);
  }) : customElements.define(t, e);
};
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const Be = { attribute: !0, type: String, converter: J, reflect: !1, hasChanged: ae }, Fe = (t = Be, e, s) => {
  const { kind: i, metadata: r } = s;
  let o = globalThis.litPropertyMetadata.get(r);
  if (o === void 0 && globalThis.litPropertyMetadata.set(r, o = /* @__PURE__ */ new Map()), i === "setter" && ((t = Object.create(t)).wrapped = !0), o.set(s.name, t), i === "accessor") {
    const { name: a } = s;
    return { set(d) {
      const n = e.get.call(this);
      e.set.call(this, d), this.requestUpdate(a, n, t, !0, d);
    }, init(d) {
      return d !== void 0 && this.C(a, void 0, t, d), d;
    } };
  }
  if (i === "setter") {
    const { name: a } = s;
    return function(d) {
      const n = this[a];
      e.call(this, d), this.requestUpdate(a, n, t, !0, d);
    };
  }
  throw Error("Unsupported decorator location: " + i);
};
function w(t) {
  return (e, s) => typeof s == "object" ? Fe(t, e, s) : ((i, r, o) => {
    const a = r.hasOwnProperty(o);
    return r.constructor.createProperty(o, i), a ? Object.getOwnPropertyDescriptor(r, o) : void 0;
  })(t, e, s);
}
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
function m(t) {
  return w({ ...t, state: !0, attribute: !1 });
}
function Ve(t, e) {
  const s = new WebSocket(t);
  return s.onmessage = (i) => {
    var r, o, a, d;
    try {
      const n = JSON.parse(i.data);
      ((o = (r = n.type) == null ? void 0 : r.startsWith) != null && o.call(r, "scm.") || (d = (a = n.channel) == null ? void 0 : a.startsWith) != null && d.call(a, "scm.")) && e(n);
    } catch {
    }
  }, s;
}
class G {
  constructor(e = "") {
    this.baseUrl = e;
  }
  get base() {
    return `${this.baseUrl}/api/v1/scm`;
  }
  async request(e, s) {
    var o, a;
    const i = await fetch(`${this.base}${e}`, s), r = await i.json().catch(() => null);
    if (!i.ok)
      throw new Error(((o = r == null ? void 0 : r.error) == null ? void 0 : o.message) ?? `Request failed (${i.status})`);
    if (!(r != null && r.success))
      throw new Error(((a = r == null ? void 0 : r.error) == null ? void 0 : a.message) ?? "Request failed");
    return r.data;
  }
  marketplace(e, s) {
    const i = new URLSearchParams();
    e && i.set("q", e), s && i.set("category", s);
    const r = i.toString();
    return this.request(`/marketplace${r ? `?${r}` : ""}`);
  }
  marketplaceItem(e) {
    return this.request(`/marketplace/${encodeURIComponent(e)}`);
  }
  install(e) {
    return this.request(`/marketplace/${encodeURIComponent(e)}/install`, { method: "POST" });
  }
  remove(e) {
    return this.request(`/marketplace/${encodeURIComponent(e)}`, { method: "DELETE" });
  }
  installed() {
    return this.request("/installed");
  }
  updateInstalled(e) {
    return this.request(`/installed/${encodeURIComponent(e)}/update`, { method: "POST" });
  }
  manifest() {
    return this.request("/manifest");
  }
  verify(e) {
    return this.request("/manifest/verify", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ public_key: e })
    });
  }
  sign(e) {
    return this.request("/manifest/sign", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ private_key: e })
    });
  }
  permissions() {
    return this.request("/manifest/permissions");
  }
  registry() {
    return this.request("/registry");
  }
}
var Ye = Object.defineProperty, Ke = Object.getOwnPropertyDescriptor, $ = (t, e, s, i) => {
  for (var r = i > 1 ? void 0 : i ? Ke(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (r = (i ? a(e, s, r) : a(r)) || r);
  return i && r && Ye(e, s, r), r;
};
let g = class extends x {
  constructor() {
    super(...arguments), this.apiUrl = "", this.category = "", this.modules = [], this.categories = [], this.searchQuery = "", this.activeCategory = "", this.loading = !0, this.error = "", this.installing = /* @__PURE__ */ new Set();
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new G(this.apiUrl), this.activeCategory = this.category, this.loadModules();
  }
  async loadModules() {
    this.loading = !0, this.error = "";
    try {
      this.modules = await this.api.marketplace(
        this.searchQuery || void 0,
        this.activeCategory || void 0
      );
      const t = /* @__PURE__ */ new Set();
      this.modules.forEach((e) => {
        e.category && t.add(e.category);
      }), this.categories = Array.from(t).sort();
    } catch (t) {
      this.error = t.message ?? "Failed to load marketplace";
    } finally {
      this.loading = !1;
    }
  }
  async refresh() {
    await this.loadModules();
  }
  handleSearch(t) {
    this.searchQuery = t.target.value, this.loadModules();
  }
  handleCategoryClick(t) {
    this.activeCategory = this.activeCategory === t ? "" : t, this.loadModules();
  }
  async handleInstall(t) {
    this.installing = /* @__PURE__ */ new Set([...this.installing, t]);
    try {
      await this.api.install(t), this.dispatchEvent(
        new CustomEvent("scm-installed", { detail: { code: t }, bubbles: !0 })
      );
    } catch (e) {
      this.error = e.message ?? "Installation failed";
    } finally {
      const e = new Set(this.installing);
      e.delete(t), this.installing = e;
    }
  }
  async handleRemove(t) {
    try {
      await this.api.remove(t), this.dispatchEvent(
        new CustomEvent("scm-removed", { detail: { code: t }, bubbles: !0 })
      );
    } catch (e) {
      this.error = e.message ?? "Removal failed";
    }
  }
  render() {
    return l`
      <div class="shell">
        <div class="summary">
          <div class="summary-copy">
            <span class="summary-title">Marketplace</span>
            <span class="summary-subtitle">
              Browse, filter, and install providers from the current index.
            </span>
          </div>
          <div class="summary-stats">
            <div class="stat">
              <span class="stat-value">${this.modules.length}</span>
              <span class="stat-label">Results</span>
            </div>
            <div class="stat">
              <span class="stat-value">${this.categories.length}</span>
              <span class="stat-label">Categories</span>
            </div>
          </div>
        </div>

        <div class="toolbar">
          <input
            type="text"
            class="search"
            placeholder="Search providers\u2026"
            .value=${this.searchQuery}
            @input=${this.handleSearch}
          />
          ${this.categories.length > 0 ? l`
                <div class="categories">
                  ${this.categories.map(
      (t) => l`
                      <button
                        class="category-btn ${this.activeCategory === t ? "active" : ""}"
                        @click=${() => this.handleCategoryClick(t)}
                      >
                        ${t}
                      </button>
                    `
    )}
                </div>
              ` : c}
        </div>

        ${this.error ? l`<div class="error">${this.error}</div>` : c}
        ${this.loading ? l`<div class="loading">Loading marketplace\u2026</div>` : this.modules.length === 0 ? l`<div class="empty">No providers found.</div>` : l`
                <div class="grid">
                  ${this.modules.map(
      (t) => l`
                      <div class="card">
                        <div>
                          <div class="card-header">
                            <div>
                              <div class="card-name">${t.name}</div>
                              <div class="card-code">${t.code}</div>
                            </div>
                            ${t.category ? l`<span class="card-category">${t.category}</span>` : c}
                          </div>
                          <div class="card-repo">${t.repo}</div>
                          <div class="card-sign ${t.sign_key ? "" : "unsigned"}">
                            ${t.sign_key ? "Signed module" : "Unsigned module"}
                          </div>
                        </div>
                        <div class="card-actions">
                          <button
                            class="install"
                            ?disabled=${this.installing.has(t.code)}
                            @click=${() => this.handleInstall(t.code)}
                          >
                            ${this.installing.has(t.code) ? "Installing…" : "Install"}
                          </button>
                          <button class="remove" @click=${() => this.handleRemove(t.code)}>
                            Remove
                          </button>
                        </div>
                      </div>
                    `
    )}
                </div>
              `}
      </div>
    `;
  }
};
g.styles = B`
    :host {
      display: block;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      color: #111827;
    }

    .shell {
      background: rgba(255, 255, 255, 0.84);
      border: 1px solid rgba(226, 232, 240, 0.95);
      border-radius: 1rem;
      padding: 1rem;
      box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
      backdrop-filter: blur(12px);
    }

    .summary {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      margin-bottom: 1rem;
      padding-bottom: 0.75rem;
      border-bottom: 1px solid #e2e8f0;
    }

    .summary-copy {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .summary-title {
      font-size: 1rem;
      font-weight: 800;
      color: #0f172a;
    }

    .summary-subtitle {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .summary-stats {
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
      justify-content: flex-end;
    }

    .stat {
      min-width: 4.5rem;
      padding: 0.55rem 0.75rem;
      border-radius: 0.85rem;
      background: linear-gradient(180deg, #f8fafc, #eef2ff);
      border: 1px solid #e2e8f0;
      text-align: center;
    }

    .stat-value {
      display: block;
      font-size: 1rem;
      font-weight: 800;
      color: #312e81;
      line-height: 1;
    }

    .stat-label {
      display: block;
      margin-top: 0.25rem;
      font-size: 0.6875rem;
      letter-spacing: 0.06em;
      text-transform: uppercase;
      color: #64748b;
    }

    .toolbar {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-bottom: 1rem;
      flex-wrap: wrap;
    }

    .search {
      flex: 1;
      min-width: 14rem;
      padding: 0.7rem 0.9rem;
      border: 1px solid #cbd5e1;
      border-radius: 0.85rem;
      font-size: 0.875rem;
      outline: none;
      background: #fff;
      transition:
        border-color 0.15s ease,
        box-shadow 0.15s ease;
    }

    .search:focus {
      border-color: #6366f1;
      box-shadow: 0 0 0 4px rgba(99, 102, 241, 0.12);
    }

    .categories {
      display: flex;
      gap: 0.25rem;
      flex-wrap: wrap;
    }

    .category-btn {
      padding: 0.35rem 0.8rem;
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      background: #fff;
      font-size: 0.75rem;
      font-weight: 700;
      color: #475569;
      cursor: pointer;
      transition:
        background 0.15s ease,
        color 0.15s ease,
        border-color 0.15s ease,
        transform 0.15s ease;
    }

    .category-btn:hover {
      background: #f8fafc;
      transform: translateY(-1px);
    }

    .category-btn.active {
      background: linear-gradient(180deg, #6366f1, #4f46e5);
      color: #fff;
      border-color: #4f46e5;
      box-shadow: 0 6px 16px rgba(99, 102, 241, 0.22);
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
      gap: 0.875rem;
      margin-top: 1rem;
    }

    .card {
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      padding: 1rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
      min-height: 10rem;
      display: flex;
      flex-direction: column;
      justify-content: space-between;
    }

    .card:hover {
      box-shadow: 0 16px 36px rgba(15, 23, 42, 0.08);
      border-color: #c7d2fe;
      transform: translateY(-2px);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 0.75rem;
      margin-bottom: 0.75rem;
    }

    .card-name {
      font-weight: 800;
      font-size: 0.95rem;
      color: #0f172a;
      line-height: 1.2;
    }

    .card-code {
      margin-top: 0.15rem;
      font-size: 0.775rem;
      color: #64748b;
      font-family: monospace;
    }

    .card-category {
      font-size: 0.6875rem;
      padding: 0.2rem 0.5rem;
      background: #e0e7ff;
      border-radius: 999px;
      color: #4338ca;
      font-weight: 700;
      white-space: nowrap;
    }

    .card-repo {
      font-size: 0.7875rem;
      color: #475569;
      font-family: monospace;
      word-break: break-word;
      margin-bottom: 0.4rem;
    }

    .card-sign {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      font-size: 0.6875rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #166534;
      margin-bottom: 0.4rem;
    }

    .card-sign::before {
      content: '';
      width: 0.45rem;
      height: 0.45rem;
      border-radius: 999px;
      background: #22c55e;
    }

    .card-sign.unsigned {
      color: #64748b;
    }

    .card-sign.unsigned::before {
      background: #f59e0b;
    }

    .card-actions {
      margin-top: 0.75rem;
      display: flex;
      gap: 0.5rem;
    }

    button.install {
      padding: 0.45rem 1rem;
      background: linear-gradient(180deg, #6366f1, #4f46e5);
      color: #fff;
      border: none;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      font-weight: 700;
      cursor: pointer;
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        background 0.15s ease;
      box-shadow: 0 8px 16px rgba(99, 102, 241, 0.2);
    }

    button.install:hover {
      background: linear-gradient(180deg, #4f46e5, #4338ca);
      transform: translateY(-1px);
    }

    button.install:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      padding: 0.45rem 1rem;
      background: #fff;
      color: #dc2626;
      border: 1px solid #dc2626;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      font-weight: 700;
      cursor: pointer;
      transition:
        background 0.15s ease,
        transform 0.15s ease;
    }

    button.remove:hover {
      background: #fef2f2;
      transform: translateY(-1px);
    }

    .empty {
      text-align: center;
      padding: 2rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #64748b;
    }

    .error {
      color: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 0.75rem;
      font-size: 0.875rem;
    }

    @media (max-width: 720px) {
      .shell {
        padding: 0.875rem;
        border-radius: 0.875rem;
      }

      .summary {
        flex-direction: column;
        align-items: flex-start;
      }

      .summary-stats {
        justify-content: flex-start;
      }

      .toolbar {
        flex-direction: column;
        align-items: stretch;
      }
    }
  `;
$([
  w({ attribute: "api-url" })
], g.prototype, "apiUrl", 2);
$([
  w()
], g.prototype, "category", 2);
$([
  m()
], g.prototype, "modules", 2);
$([
  m()
], g.prototype, "categories", 2);
$([
  m()
], g.prototype, "searchQuery", 2);
$([
  m()
], g.prototype, "activeCategory", 2);
$([
  m()
], g.prototype, "loading", 2);
$([
  m()
], g.prototype, "error", 2);
$([
  m()
], g.prototype, "installing", 2);
g = $([
  V("core-scm-marketplace")
], g);
var Je = Object.defineProperty, Qe = Object.getOwnPropertyDescriptor, I = (t, e, s, i) => {
  for (var r = i > 1 ? void 0 : i ? Qe(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (r = (i ? a(e, s, r) : a(r)) || r);
  return i && r && Je(e, s, r), r;
};
let A = class extends x {
  constructor() {
    super(...arguments), this.apiUrl = "", this.modules = [], this.loading = !0, this.error = "", this.updating = /* @__PURE__ */ new Set();
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new G(this.apiUrl), this.loadInstalled();
  }
  async loadInstalled() {
    this.loading = !0, this.error = "";
    try {
      this.modules = await this.api.installed();
    } catch (t) {
      this.error = t.message ?? "Failed to load installed providers";
    } finally {
      this.loading = !1;
    }
  }
  async refresh() {
    await this.loadInstalled();
  }
  async handleUpdate(t) {
    this.updating = /* @__PURE__ */ new Set([...this.updating, t]);
    try {
      await this.api.updateInstalled(t), await this.loadInstalled();
    } catch (e) {
      this.error = e.message ?? "Update failed";
    } finally {
      const e = new Set(this.updating);
      e.delete(t), this.updating = e;
    }
  }
  async handleRemove(t) {
    try {
      await this.api.remove(t), this.dispatchEvent(
        new CustomEvent("scm-removed", { detail: { code: t }, bubbles: !0 })
      ), await this.loadInstalled();
    } catch (e) {
      this.error = e.message ?? "Removal failed";
    }
  }
  formatDate(t) {
    try {
      return new Date(t).toLocaleDateString("en-GB", {
        day: "numeric",
        month: "short",
        year: "numeric"
      });
    } catch {
      return t;
    }
  }
  render() {
    return this.loading ? l`<div class="loading">Loading installed providers\u2026</div>` : l`
      <div class="shell">
        <div class="summary">
          <div class="summary-copy">
            <span class="summary-title">Installed providers</span>
            <span class="summary-subtitle">
              Review local modules, update them, or remove stale installs.
            </span>
          </div>
          <div class="summary-copy" style="text-align:right">
            <span class="summary-title">${this.modules.length}</span>
            <span class="summary-subtitle">Installed</span>
          </div>
        </div>

        ${this.error ? l`<div class="error">${this.error}</div>` : c}
        ${this.modules.length === 0 ? l`<div class="empty">No providers installed.</div>` : l`
              <div class="list">
                ${this.modules.map(
      (t) => l`
                    <div class="item">
                      <div class="item-info">
                        <div class="item-name">${t.name}</div>
                        <div class="item-meta">
                          <span class="item-code">${t.code}</span>
                          <span>v${t.version}</span>
                          <span>Installed ${this.formatDate(t.installed_at)}</span>
                        </div>
                        <div class="item-meta">
                          <span class="item-repo">${t.repo}</span>
                          <span class="item-entry">entry: ${t.entry_point}</span>
                        </div>
                        <div class="badge ${t.sign_key ? "" : "unsigned"}">
                          ${t.sign_key ? "Signed manifest" : "Unsigned manifest"}
                        </div>
                      </div>
                      <div class="item-actions">
                        <button
                          class="update"
                          ?disabled=${this.updating.has(t.code)}
                          @click=${() => this.handleUpdate(t.code)}
                        >
                          ${this.updating.has(t.code) ? "Updating…" : "Update"}
                        </button>
                        <button class="remove" @click=${() => this.handleRemove(t.code)}>
                          Remove
                        </button>
                      </div>
                    </div>
                  `
    )}
              </div>
            `}
      </div>
    `;
  }
};
A.styles = B`
    :host {
      display: block;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      color: #111827;
    }

    .shell {
      background: rgba(255, 255, 255, 0.84);
      border: 1px solid rgba(226, 232, 240, 0.95);
      border-radius: 1rem;
      padding: 1rem;
      box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
      backdrop-filter: blur(12px);
    }

    .summary {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      margin-bottom: 1rem;
      padding-bottom: 0.75rem;
      border-bottom: 1px solid #e2e8f0;
    }

    .summary-copy {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .summary-title {
      font-size: 1rem;
      font-weight: 800;
      color: #0f172a;
    }

    .summary-subtitle {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.75rem;
    }

    .item {
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      padding: 1rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        border-color 0.15s ease;
    }

    .item:hover {
      box-shadow: 0 16px 36px rgba(15, 23, 42, 0.08);
      border-color: #c7d2fe;
      transform: translateY(-2px);
    }

    .item-info {
      flex: 1;
    }

    .item-name {
      font-weight: 800;
      font-size: 0.95rem;
      color: #0f172a;
    }

    .item-meta {
      font-size: 0.75rem;
      color: #64748b;
      margin-top: 0.35rem;
      display: flex;
      gap: 1rem;
      flex-wrap: wrap;
    }

    .item-code {
      font-family: monospace;
      color: #334155;
      font-weight: 700;
    }

    .item-repo,
    .item-entry {
      font-family: monospace;
      color: #475569;
      word-break: break-word;
    }

    .badge {
      display: inline-flex;
      align-items: center;
      gap: 0.3rem;
      font-size: 0.6875rem;
      font-weight: 700;
      text-transform: uppercase;
      letter-spacing: 0.05em;
      color: #166534;
    }

    .badge::before {
      content: '';
      width: 0.45rem;
      height: 0.45rem;
      border-radius: 999px;
      background: #22c55e;
    }

    .badge.unsigned {
      color: #64748b;
    }

    .badge.unsigned::before {
      background: #f59e0b;
    }

    .item-actions {
      display: flex;
      gap: 0.5rem;
    }

    button {
      padding: 0.45rem 0.85rem;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      cursor: pointer;
      font-weight: 700;
      transition:
        background 0.15s ease,
        transform 0.15s ease,
        box-shadow 0.15s ease;
    }

    button.update {
      background: #fff;
      color: #4338ca;
      border: 1px solid #6366f1;
    }

    button.update:hover {
      background: #eef2ff;
      transform: translateY(-1px);
    }

    button.update:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      background: #fff;
      color: #dc2626;
      border: 1px solid #dc2626;
    }

    button.remove:hover {
      background: #fef2f2;
      transform: translateY(-1px);
    }

    .empty {
      text-align: center;
      padding: 2rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #64748b;
    }

    .error {
      color: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }

    @media (max-width: 720px) {
      .shell {
        padding: 0.875rem;
      }

      .summary,
      .item {
        flex-direction: column;
        align-items: flex-start;
      }

      .item-actions {
        width: 100%;
      }

      button {
        flex: 1;
      }
    }
  `;
I([
  w({ attribute: "api-url" })
], A.prototype, "apiUrl", 2);
I([
  m()
], A.prototype, "modules", 2);
I([
  m()
], A.prototype, "loading", 2);
I([
  m()
], A.prototype, "error", 2);
I([
  m()
], A.prototype, "updating", 2);
A = I([
  V("core-scm-installed")
], A);
var Ze = Object.defineProperty, Ge = Object.getOwnPropertyDescriptor, C = (t, e, s, i) => {
  for (var r = i > 1 ? void 0 : i ? Ge(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (r = (i ? a(e, s, r) : a(r)) || r);
  return i && r && Ze(e, s, r), r;
};
let v = class extends x {
  constructor() {
    super(...arguments), this.apiUrl = "", this.path = "", this.manifest = null, this.loading = !0, this.error = "", this.verifyKey = "", this.verifyResult = null;
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new G(this.apiUrl), this.loadManifest();
  }
  async loadManifest() {
    this.loading = !0, this.error = "";
    try {
      this.manifest = await this.api.manifest();
    } catch (t) {
      this.error = t.message ?? "Failed to load manifest";
    } finally {
      this.loading = !1;
    }
  }
  async refresh() {
    this.verifyResult = null, await this.loadManifest();
  }
  async handleVerify() {
    if (this.verifyKey.trim())
      try {
        this.verifyResult = await this.api.verify(this.verifyKey.trim());
      } catch (t) {
        this.error = t.message ?? "Verification failed";
      }
  }
  async handleSign() {
    const t = prompt("Enter hex-encoded Ed25519 private key:");
    if (t)
      try {
        await this.api.sign(t), await this.loadManifest();
      } catch (e) {
        this.error = e.message ?? "Signing failed";
      }
  }
  renderPermissions(t) {
    if (!t) return c;
    const e = [
      { label: "Read", items: t.read },
      { label: "Write", items: t.write },
      { label: "Network", items: t.net },
      { label: "Run", items: t.run }
    ].filter((s) => s.items && s.items.length > 0);
    return e.length === 0 ? c : l`
      <div class="field">
        <div class="field-label">Permissions</div>
        <div class="permissions-grid">
          ${e.map(
      (s) => l`
              <div class="perm-group">
                <div class="perm-group-label">${s.label}</div>
                ${s.items.map((i) => l`<div class="perm-item">${i}</div>`)}
              </div>
            `
    )}
        </div>
      </div>
    `;
  }
  render() {
    var o, a, d, n, h, u, p, f, b;
    if (this.loading)
      return l`<div class="loading">Loading manifest\u2026</div>`;
    if (this.error && !this.manifest)
      return l`<div class="error">${this.error}</div>`;
    if (!this.manifest)
      return l`<div class="empty">No manifest found. Create a .core/manifest.yaml to get started.</div>`;
    const t = this.manifest, e = !!t.sign, s = e ? this.verifyResult ? this.verifyResult.valid ? "verified" : "invalid" : "signed" : "unsigned", i = e ? this.verifyResult ? this.verifyResult.valid ? "Verified" : "Invalid" : "Signed" : "Unsigned", r = (((a = (o = t.permissions) == null ? void 0 : o.read) == null ? void 0 : a.length) ?? 0) + (((n = (d = t.permissions) == null ? void 0 : d.write) == null ? void 0 : n.length) ?? 0) + (((u = (h = t.permissions) == null ? void 0 : h.net) == null ? void 0 : u.length) ?? 0) + (((f = (p = t.permissions) == null ? void 0 : p.run) == null ? void 0 : f.length) ?? 0);
    return l`
      ${this.error ? l`<div class="error">${this.error}</div>` : c}
      <div class="manifest">
        <div class="header">
          <div class="header-copy">
            <h3>${t.name}</h3>
            <div class="meta-row">
              <span class="meta-chip">${t.code}</span>
              <span class="meta-chip">${r} permissions</span>
              ${(b = t.modules) != null && b.length ? l`<span class="meta-chip">${t.modules.length} modules</span>` : c}
            </div>
          </div>
          <span class="version">v${t.version}</span>
        </div>

        ${t.description ? l`
              <div class="field">
                <div class="field-label">Description</div>
                <div class="field-value">${t.description}</div>
              </div>
            ` : c}
        ${t.layout ? l`
              <div class="field">
                <div class="field-label">Layout</div>
                <div class="field-value code">${t.layout}</div>
              </div>
            ` : c}
        ${t.slots && Object.keys(t.slots).length > 0 ? l`
              <div class="field">
                <div class="field-label">Slots</div>
                <div class="slots">
                  ${Object.entries(t.slots).map(
      ([y, ke]) => l`
                      <span class="slot-key">${y}</span>
                      <span class="slot-value">${ke}</span>
                    `
    )}
                </div>
              </div>
            ` : c}

        ${this.renderPermissions(t.permissions)}
        ${t.modules && t.modules.length > 0 ? l`
              <div class="field">
                <div class="field-label">Modules</div>
                ${t.modules.map((y) => l`<div class="code" style="margin-bottom:0.35rem">${y}</div>`)}
              </div>
            ` : c}

        <div class="signature ${s}">
          <span class="badge ${s}">${i}</span>
          <span class="status">
            <span class="status-dot ${s}"></span>
            ${e ? l`<span>Signature present</span>` : l`<span>No signature</span>`}
          </span>
        </div>

        <div class="actions">
          <input
            type="text"
            class="verify-input"
            placeholder="Public key (hex)\u2026"
            .value=${this.verifyKey}
            @input=${(y) => this.verifyKey = y.target.value}
          />
          <button @click=${this.handleVerify}>Verify</button>
          <button class="primary" @click=${this.handleSign}>Sign</button>
        </div>
      </div>
    `;
  }
};
v.styles = B`
    :host {
      display: block;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      color: #111827;
    }

    .manifest {
      border: 1px solid rgba(226, 232, 240, 0.95);
      border-radius: 1rem;
      padding: 1.25rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
    }

    .header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      gap: 1rem;
      margin-bottom: 1rem;
      padding-bottom: 0.9rem;
      border-bottom: 1px solid #e2e8f0;
    }

    .header-copy {
      display: flex;
      flex-direction: column;
      gap: 0.35rem;
    }

    h3 {
      margin: 0;
      font-size: 1.2rem;
      font-weight: 800;
      color: #0f172a;
    }

    .version {
      font-size: 0.75rem;
      padding: 0.25rem 0.625rem;
      background: #eef2ff;
      border-radius: 999px;
      color: #4338ca;
      font-weight: 700;
    }

    .meta-row {
      display: flex;
      flex-wrap: wrap;
      gap: 0.5rem;
    }

    .meta-chip {
      display: inline-flex;
      align-items: center;
      gap: 0.35rem;
      padding: 0.25rem 0.6rem;
      border-radius: 999px;
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      font-size: 0.6875rem;
      font-weight: 700;
      color: #475569;
      text-transform: uppercase;
      letter-spacing: 0.05em;
    }

    .field {
      margin-bottom: 0.875rem;
    }

    .field-label {
      font-size: 0.75rem;
      font-weight: 800;
      color: #64748b;
      text-transform: uppercase;
      letter-spacing: 0.025em;
      margin-bottom: 0.25rem;
    }

    .field-value {
      font-size: 0.875rem;
      color: #0f172a;
    }

    .code {
      font-family: monospace;
      font-size: 0.8125rem;
      background: #f8fafc;
      padding: 0.35rem 0.55rem;
      border-radius: 0.45rem;
      border: 1px solid #e2e8f0;
      color: #1f2937;
    }

    .slots {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 0.25rem 1rem;
      font-size: 0.8125rem;
      background: #f8fafc;
      border: 1px solid #e2e8f0;
      border-radius: 0.75rem;
      padding: 0.75rem;
    }

    .slot-key {
      font-weight: 700;
      color: #334155;
    }

    .slot-value {
      font-family: monospace;
      color: #64748b;
    }

    .permissions-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 0.5rem;
    }

    .perm-group {
      border: 1px solid #e2e8f0;
      border-radius: 0.75rem;
      padding: 0.5rem;
      background: #fff;
    }

    .perm-group-label {
      font-size: 0.6875rem;
      font-weight: 800;
      color: #64748b;
      text-transform: uppercase;
      margin-bottom: 0.25rem;
    }

    .perm-item {
      font-size: 0.8125rem;
      font-family: monospace;
      color: #374151;
      word-break: break-word;
    }

    .signature {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-top: 1rem;
      padding: 0.75rem;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      font-weight: 600;
    }

    .signature.signed {
      background: #f0fdf4;
      border: 1px solid #bbf7d0;
    }

    .signature.unsigned {
      background: #fffbeb;
      border: 1px solid #fde68a;
    }

    .signature.invalid {
      background: #fef2f2;
      border: 1px solid #fecaca;
    }

    .badge {
      font-size: 0.75rem;
      font-weight: 800;
      padding: 0.25rem 0.55rem;
      border-radius: 999px;
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }

    .badge.verified {
      background: #dcfce7;
      color: #166534;
    }

    .badge.signed {
      background: #e0e7ff;
      color: #4338ca;
    }

    .badge.unsigned {
      background: #fef3c7;
      color: #92400e;
    }

    .badge.invalid {
      background: #fee2e2;
      color: #991b1b;
    }

    .actions {
      margin-top: 1rem;
      display: flex;
      gap: 0.5rem;
      flex-wrap: wrap;
    }

    .verify-input {
      flex: 1;
      min-width: 14rem;
      padding: 0.6rem 0.85rem;
      border: 1px solid #cbd5e1;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      font-family: monospace;
      background: #fff;
    }

    button {
      padding: 0.6rem 1rem;
      border-radius: 0.75rem;
      font-size: 0.8125rem;
      cursor: pointer;
      border: 1px solid #d1d5db;
      background: #fff;
      transition:
        background 0.15s ease,
        transform 0.15s ease;
    }

    button:hover {
      background: #f3f4f6;
      transform: translateY(-1px);
    }

    button.primary {
      background: linear-gradient(180deg, #6366f1, #4f46e5);
      color: #fff;
      border-color: #6366f1;
    }

    button.primary:hover {
      background: linear-gradient(180deg, #4f46e5, #4338ca);
    }

    .empty {
      text-align: center;
      padding: 2rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    .error {
      color: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      margin-bottom: 0.75rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #64748b;
    }

    .status {
      display: flex;
      align-items: center;
      gap: 0.35rem;
    }

    .status-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 999px;
    }

    .status-dot.verified {
      background: #22c55e;
    }

    .status-dot.signed {
      background: #6366f1;
    }

    .status-dot.unsigned {
      background: #f59e0b;
    }

    .status-dot.invalid {
      background: #ef4444;
    }

    @media (max-width: 720px) {
      .manifest {
        padding: 1rem;
      }

      .header {
        flex-direction: column;
        align-items: flex-start;
      }

      .actions {
        flex-direction: column;
      }
    }
  `;
C([
  w({ attribute: "api-url" })
], v.prototype, "apiUrl", 2);
C([
  w()
], v.prototype, "path", 2);
C([
  m()
], v.prototype, "manifest", 2);
C([
  m()
], v.prototype, "loading", 2);
C([
  m()
], v.prototype, "error", 2);
C([
  m()
], v.prototype, "verifyKey", 2);
C([
  m()
], v.prototype, "verifyResult", 2);
v = C([
  V("core-scm-manifest")
], v);
var Xe = Object.defineProperty, et = Object.getOwnPropertyDescriptor, Y = (t, e, s, i) => {
  for (var r = i > 1 ? void 0 : i ? et(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (r = (i ? a(e, s, r) : a(r)) || r);
  return i && r && Xe(e, s, r), r;
};
let z = class extends x {
  constructor() {
    super(...arguments), this.apiUrl = "", this.repos = [], this.loading = !0, this.error = "";
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new G(this.apiUrl), this.loadRegistry();
  }
  async loadRegistry() {
    this.loading = !0, this.error = "";
    try {
      this.repos = await this.api.registry();
    } catch (t) {
      this.error = t.message ?? "Failed to load registry";
    } finally {
      this.loading = !1;
    }
  }
  async refresh() {
    await this.loadRegistry();
  }
  render() {
    return this.loading ? l`<div class="loading">Loading registry\u2026</div>` : l`
      <div class="shell">
        <div class="summary">
          <div class="summary-copy">
            <span class="summary-title">Registry</span>
            <span class="summary-subtitle">
              Workspace repositories and dependency order from repos.yaml.
            </span>
          </div>
          <div class="summary-copy" style="text-align:right">
            <span class="summary-title">${this.repos.length}</span>
            <span class="summary-subtitle">Entries</span>
          </div>
        </div>

        ${this.error ? l`<div class="error">${this.error}</div>` : c}
        ${this.repos.length === 0 ? l`<div class="empty">No repositories in registry.</div>` : l`
              <div class="list">
                ${this.repos.map(
      (t) => l`
                    <div class="repo">
                      <div class="repo-info">
                        <div class="repo-name">${t.name}</div>
                        ${t.description ? l`<div class="repo-desc">${t.description}</div>` : c}
                        <div class="repo-meta">
                          <span class="type-badge ${t.type}">${t.type}</span>
                          ${t.depends_on && t.depends_on.length > 0 ? l`<span class="deps">depends: ${t.depends_on.join(", ")}</span>` : c}
                        </div>
                        ${t.path ? l`<div class="path">${t.path}</div>` : c}
                      </div>
                      <div class="status">
                        <span class="status-dot ${t.exists ? "present" : "missing"}"></span>
                        <span class="status-label">${t.exists ? "Present" : "Missing"}</span>
                      </div>
                    </div>
                  `
    )}
              </div>
            `}
      </div>
    `;
  }
};
z.styles = B`
    :host {
      display: block;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      color: #111827;
    }

    .shell {
      background: rgba(255, 255, 255, 0.84);
      border: 1px solid rgba(226, 232, 240, 0.95);
      border-radius: 1rem;
      padding: 1rem;
      box-shadow: 0 18px 40px rgba(15, 23, 42, 0.06);
      backdrop-filter: blur(12px);
    }

    .summary {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      margin-bottom: 1rem;
      padding-bottom: 0.75rem;
      border-bottom: 1px solid #e2e8f0;
    }

    .summary-copy {
      display: flex;
      flex-direction: column;
      gap: 0.25rem;
    }

    .summary-title {
      font-size: 1rem;
      font-weight: 800;
      color: #0f172a;
    }

    .summary-subtitle {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.625rem;
    }

    .repo {
      border: 1px solid #e2e8f0;
      border-radius: 1rem;
      padding: 0.75rem 1rem;
      background:
        linear-gradient(180deg, rgba(255, 255, 255, 0.98), rgba(248, 250, 252, 0.98));
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
    }

    .repo-info {
      flex: 1;
    }

    .repo-name {
      font-weight: 800;
      font-size: 0.95rem;
      font-family: monospace;
      color: #0f172a;
    }

    .repo-desc {
      font-size: 0.8125rem;
      color: #64748b;
      margin-top: 0.125rem;
    }

    .repo-meta {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-top: 0.25rem;
      flex-wrap: wrap;
    }

    .type-badge {
      font-size: 0.6875rem;
      padding: 0.2rem 0.5rem;
      border-radius: 999px;
      font-weight: 800;
      text-transform: uppercase;
      letter-spacing: 0.04em;
    }

    .type-badge.foundation {
      background: #dbeafe;
      color: #1e40af;
    }

    .type-badge.module {
      background: #f3e8ff;
      color: #6b21a8;
    }

    .type-badge.product {
      background: #dcfce7;
      color: #166534;
    }

    .type-badge.template {
      background: #fef3c7;
      color: #92400e;
    }

    .deps {
      font-size: 0.75rem;
      color: #64748b;
    }

    .path {
      font-size: 0.75rem;
      font-family: monospace;
      color: #475569;
      word-break: break-word;
    }

    .status {
      display: flex;
      align-items: center;
      gap: 0.375rem;
    }

    .status-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 50%;
    }

    .status-dot.present {
      background: #22c55e;
      box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.15);
    }

    .status-dot.missing {
      background: #ef4444;
      box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.14);
    }

    .status-label {
      font-size: 0.75rem;
      color: #64748b;
      font-weight: 700;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      color: #64748b;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      color: #64748b;
    }

    .error {
      color: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border: 1px solid #fecaca;
      border-radius: 0.75rem;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }

    @media (max-width: 720px) {
      .shell {
        padding: 0.875rem;
      }

      .summary,
      .repo {
        flex-direction: column;
        align-items: flex-start;
      }
    }
  `;
Y([
  w({ attribute: "api-url" })
], z.prototype, "apiUrl", 2);
Y([
  m()
], z.prototype, "repos", 2);
Y([
  m()
], z.prototype, "loading", 2);
Y([
  m()
], z.prototype, "error", 2);
z = Y([
  V("core-scm-registry")
], z);
var tt = Object.defineProperty, st = Object.getOwnPropertyDescriptor, N = (t, e, s, i) => {
  for (var r = i > 1 ? void 0 : i ? st(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (r = (i ? a(e, s, r) : a(r)) || r);
  return i && r && tt(e, s, r), r;
};
let S = class extends x {
  constructor() {
    super(...arguments), this.apiUrl = "", this.wsUrl = "", this.activeTab = "marketplace", this.wsConnected = !1, this.lastEvent = "", this.ws = null, this.tabs = [
      { id: "marketplace", label: "Marketplace" },
      { id: "installed", label: "Installed" },
      { id: "manifest", label: "Manifest" },
      { id: "registry", label: "Registry" }
    ];
  }
  connectedCallback() {
    super.connectedCallback(), this.wsUrl && this.connectWs();
  }
  updated(t) {
    super.updated(t), t.has("wsUrl") && this.isConnected && this.connectWs();
  }
  disconnectedCallback() {
    super.disconnectedCallback(), this.disconnectWs();
  }
  connectWs() {
    this.disconnectWs(), this.wsUrl && (this.ws = Ve(this.wsUrl, (t) => {
      this.lastEvent = t.channel ?? t.type ?? "", this.requestUpdate(), this.refreshForEvent(t);
    }), this.ws.onopen = () => {
      this.wsConnected = !0;
    }, this.ws.onclose = () => {
      this.wsConnected = !1;
    });
  }
  disconnectWs() {
    this.ws && (this.ws.close(), this.ws = null);
  }
  handleTabClick(t) {
    this.activeTab = t;
  }
  async handleRefresh() {
    await this.refreshActiveTab();
  }
  refreshForEvent(t) {
    this.tabsForChannel(t.channel ?? t.type ?? "").includes(this.activeTab) && this.refreshActiveTab();
  }
  tabsForChannel(t) {
    return t.startsWith("scm.marketplace.") ? ["marketplace", "installed"] : t.startsWith("scm.installed.") ? ["installed"] : t === "scm.manifest.verified" ? ["manifest"] : t === "scm.registry.changed" ? ["registry"] : [];
  }
  async refreshActiveTab() {
    var e;
    const t = (e = this.shadowRoot) == null ? void 0 : e.querySelector(".content > *");
    t != null && t.refresh && await t.refresh();
  }
  renderContent() {
    switch (this.activeTab) {
      case "marketplace":
        return l`<core-scm-marketplace api-url=${this.apiUrl}></core-scm-marketplace>`;
      case "installed":
        return l`<core-scm-installed api-url=${this.apiUrl}></core-scm-installed>`;
      case "manifest":
        return l`<core-scm-manifest api-url=${this.apiUrl}></core-scm-manifest>`;
      case "registry":
        return l`<core-scm-registry api-url=${this.apiUrl}></core-scm-registry>`;
      default:
        return c;
    }
  }
  render() {
    const t = this.wsUrl ? this.wsConnected ? "connected" : "disconnected" : "idle";
    return l`
      <div class="header">
        <div class="title-wrap">
          <span class="title">SCM</span>
          <span class="subtitle">Marketplace, manifests, installed modules, and registry status</span>
        </div>
        <button class="refresh-btn" @click=${this.handleRefresh}>Refresh</button>
      </div>

      <div class="tabs">
        ${this.tabs.map(
      (e) => l`
            <button
              class="tab ${this.activeTab === e.id ? "active" : ""}"
              @click=${() => this.handleTabClick(e.id)}
            >
              ${e.label}
            </button>
          `
    )}
      </div>

      <div class="content">${this.renderContent()}</div>

      <div class="footer">
        <div class="ws-status">
          <span class="ws-dot ${t}"></span>
          <span>${t === "connected" ? "Connected" : t === "disconnected" ? "Disconnected" : "No WebSocket"}</span>
        </div>
        ${this.lastEvent ? l`<span>Last: ${this.lastEvent}</span>` : c}
      </div>
    `;
  }
};
S.styles = B`
    :host {
      display: flex;
      flex-direction: column;
      font-family:
        Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, 'Segoe UI',
        sans-serif;
      height: 100%;
      background:
        radial-gradient(circle at top left, rgba(99, 102, 241, 0.12), transparent 30%),
        linear-gradient(180deg, #eef2ff 0%, #f8fafc 28%, #f3f4f6 100%);
      color: #111827;
    }

    /* H — Header */
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      padding: 1rem 1.25rem;
      background: rgba(255, 255, 255, 0.86);
      backdrop-filter: blur(18px);
      border-bottom: 1px solid rgba(226, 232, 240, 0.9);
    }

    .title-wrap {
      display: flex;
      flex-direction: column;
      gap: 0.125rem;
    }

    .title {
      font-size: 1rem;
      font-weight: 800;
      letter-spacing: 0.08em;
      text-transform: uppercase;
      color: #0f172a;
    }

    .subtitle {
      font-size: 0.8125rem;
      color: #64748b;
    }

    .refresh-btn {
      padding: 0.5rem 0.875rem;
      border: 1px solid rgba(99, 102, 241, 0.25);
      border-radius: 999px;
      background: linear-gradient(180deg, #ffffff, #eef2ff);
      color: #4338ca;
      font-weight: 600;
      font-size: 0.8125rem;
      cursor: pointer;
      transition:
        transform 0.15s ease,
        box-shadow 0.15s ease,
        background 0.15s ease;
      box-shadow: 0 1px 1px rgba(15, 23, 42, 0.04);
    }

    .refresh-btn:hover {
      background: linear-gradient(180deg, #ffffff, #e0e7ff);
      transform: translateY(-1px);
      box-shadow: 0 8px 20px rgba(99, 102, 241, 0.12);
    }

    /* H-L — Tabs */
    .tabs {
      display: flex;
      gap: 0.375rem;
      padding: 0.75rem 1rem 0;
      background: rgba(255, 255, 255, 0.72);
      backdrop-filter: blur(18px);
      border-bottom: 1px solid rgba(226, 232, 240, 0.9);
      overflow-x: auto;
    }

    .tab {
      padding: 0.7rem 1rem;
      font-size: 0.8125rem;
      font-weight: 700;
      letter-spacing: 0.01em;
      color: #64748b;
      cursor: pointer;
      border: 1px solid transparent;
      border-radius: 999px 999px 0 0;
      transition:
        color 0.15s ease,
        background 0.15s ease,
        border-color 0.15s ease,
        transform 0.15s ease;
      background: transparent;
    }

    .tab:hover {
      color: #334155;
      transform: translateY(-1px);
    }

    .tab.active {
      color: #4338ca;
      background: rgba(255, 255, 255, 0.96);
      border-color: rgba(226, 232, 240, 0.9);
      border-bottom-color: rgba(255, 255, 255, 0.96);
      box-shadow: 0 -1px 0 rgba(255, 255, 255, 0.6), 0 -8px 24px rgba(15, 23, 42, 0.04);
    }

    /* C — Content */
    .content {
      flex: 1;
      padding: 1.25rem;
      overflow-y: auto;
      display: flex;
      justify-content: center;
      align-items: flex-start;
    }

    .content > * {
      width: min(100%, 1120px);
    }

    /* F — Footer / Status bar */
    .footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 1rem;
      padding: 0.75rem 1.25rem;
      background: rgba(255, 255, 255, 0.84);
      backdrop-filter: blur(18px);
      border-top: 1px solid rgba(226, 232, 240, 0.9);
      font-size: 0.75rem;
      color: #64748b;
    }

    .ws-status {
      display: flex;
      align-items: center;
      gap: 0.375rem;
      font-weight: 600;
    }

    .ws-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 50%;
    }

    .ws-dot.connected {
      background: #22c55e;
      box-shadow: 0 0 0 4px rgba(34, 197, 94, 0.15);
    }

    .ws-dot.disconnected {
      background: #ef4444;
      box-shadow: 0 0 0 4px rgba(239, 68, 68, 0.14);
    }

    .ws-dot.idle {
      background: #d1d5db;
    }

    @media (max-width: 720px) {
      .header,
      .footer {
        flex-direction: column;
        align-items: flex-start;
      }

      .tabs {
        padding-inline: 0.75rem;
      }

      .content {
        padding: 0.875rem;
      }
    }
  `;
N([
  w({ attribute: "api-url" })
], S.prototype, "apiUrl", 2);
N([
  w({ attribute: "ws-url" })
], S.prototype, "wsUrl", 2);
N([
  m()
], S.prototype, "activeTab", 2);
N([
  m()
], S.prototype, "wsConnected", 2);
N([
  m()
], S.prototype, "lastEvent", 2);
S = N([
  V("core-scm-panel")
], S);
export {
  G as ScmApi,
  A as ScmInstalled,
  v as ScmManifest,
  g as ScmMarketplace,
  S as ScmPanel,
  z as ScmRegistry,
  Ve as connectScmEvents
};
