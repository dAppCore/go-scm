/**
 * @license
 * Copyright 2019 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const J = globalThis, ie = J.ShadowRoot && (J.ShadyCSS === void 0 || J.ShadyCSS.nativeShadow) && "adoptedStyleSheets" in Document.prototype && "replace" in CSSStyleSheet.prototype, re = Symbol(), ne = /* @__PURE__ */ new WeakMap();
let ye = class {
  constructor(e, s, r) {
    if (this._$cssResult$ = !0, r !== re) throw Error("CSSResult is not constructable. Use `unsafeCSS` or `css` instead.");
    this.cssText = e, this.t = s;
  }
  get styleSheet() {
    let e = this.o;
    const s = this.t;
    if (ie && e === void 0) {
      const r = s !== void 0 && s.length === 1;
      r && (e = ne.get(s)), e === void 0 && ((this.o = e = new CSSStyleSheet()).replaceSync(this.cssText), r && ne.set(s, e));
    }
    return e;
  }
  toString() {
    return this.cssText;
  }
};
const xe = (t) => new ye(typeof t == "string" ? t : t + "", void 0, re), W = (t, ...e) => {
  const s = t.length === 1 ? t[0] : e.reduce((r, i, o) => r + ((a) => {
    if (a._$cssResult$ === !0) return a.cssText;
    if (typeof a == "number") return a;
    throw Error("Value passed to 'css' function must be a 'css' function result: " + a + ". Use 'unsafeCSS' to pass non-literal values, but take care to ensure page security.");
  })(i) + t[o + 1], t[0]);
  return new ye(s, t, re);
}, Se = (t, e) => {
  if (ie) t.adoptedStyleSheets = e.map((s) => s instanceof CSSStyleSheet ? s : s.styleSheet);
  else for (const s of e) {
    const r = document.createElement("style"), i = J.litNonce;
    i !== void 0 && r.setAttribute("nonce", i), r.textContent = s.cssText, t.appendChild(r);
  }
}, le = ie ? (t) => t : (t) => t instanceof CSSStyleSheet ? ((e) => {
  let s = "";
  for (const r of e.cssRules) s += r.cssText;
  return xe(s);
})(t) : t;
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const { is: Ee, defineProperty: Ce, getOwnPropertyDescriptor: ke, getOwnPropertyNames: Pe, getOwnPropertySymbols: Ue, getPrototypeOf: Oe } = Object, A = globalThis, de = A.trustedTypes, Re = de ? de.emptyScript : "", Y = A.reactiveElementPolyfillSupport, D = (t, e) => t, Q = { toAttribute(t, e) {
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
} }, ae = (t, e) => !Ee(t, e), ce = { attribute: !0, type: String, converter: Q, reflect: !1, useDefault: !1, hasChanged: ae };
Symbol.metadata ?? (Symbol.metadata = Symbol("metadata")), A.litPropertyMetadata ?? (A.litPropertyMetadata = /* @__PURE__ */ new WeakMap());
let R = class extends HTMLElement {
  static addInitializer(e) {
    this._$Ei(), (this.l ?? (this.l = [])).push(e);
  }
  static get observedAttributes() {
    return this.finalize(), this._$Eh && [...this._$Eh.keys()];
  }
  static createProperty(e, s = ce) {
    if (s.state && (s.attribute = !1), this._$Ei(), this.prototype.hasOwnProperty(e) && ((s = Object.create(s)).wrapped = !0), this.elementProperties.set(e, s), !s.noAccessor) {
      const r = Symbol(), i = this.getPropertyDescriptor(e, r, s);
      i !== void 0 && Ce(this.prototype, e, i);
    }
  }
  static getPropertyDescriptor(e, s, r) {
    const { get: i, set: o } = ke(this.prototype, e) ?? { get() {
      return this[s];
    }, set(a) {
      this[s] = a;
    } };
    return { get: i, set(a) {
      const d = i == null ? void 0 : i.call(this);
      o == null || o.call(this, a), this.requestUpdate(e, d, r);
    }, configurable: !0, enumerable: !0 };
  }
  static getPropertyOptions(e) {
    return this.elementProperties.get(e) ?? ce;
  }
  static _$Ei() {
    if (this.hasOwnProperty(D("elementProperties"))) return;
    const e = Oe(this);
    e.finalize(), e.l !== void 0 && (this.l = [...e.l]), this.elementProperties = new Map(e.elementProperties);
  }
  static finalize() {
    if (this.hasOwnProperty(D("finalized"))) return;
    if (this.finalized = !0, this._$Ei(), this.hasOwnProperty(D("properties"))) {
      const s = this.properties, r = [...Pe(s), ...Ue(s)];
      for (const i of r) this.createProperty(i, s[i]);
    }
    const e = this[Symbol.metadata];
    if (e !== null) {
      const s = litPropertyMetadata.get(e);
      if (s !== void 0) for (const [r, i] of s) this.elementProperties.set(r, i);
    }
    this._$Eh = /* @__PURE__ */ new Map();
    for (const [s, r] of this.elementProperties) {
      const i = this._$Eu(s, r);
      i !== void 0 && this._$Eh.set(i, s);
    }
    this.elementStyles = this.finalizeStyles(this.styles);
  }
  static finalizeStyles(e) {
    const s = [];
    if (Array.isArray(e)) {
      const r = new Set(e.flat(1 / 0).reverse());
      for (const i of r) s.unshift(le(i));
    } else e !== void 0 && s.push(le(e));
    return s;
  }
  static _$Eu(e, s) {
    const r = s.attribute;
    return r === !1 ? void 0 : typeof r == "string" ? r : typeof e == "string" ? e.toLowerCase() : void 0;
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
    for (const r of s.keys()) this.hasOwnProperty(r) && (e.set(r, this[r]), delete this[r]);
    e.size > 0 && (this._$Ep = e);
  }
  createRenderRoot() {
    const e = this.shadowRoot ?? this.attachShadow(this.constructor.shadowRootOptions);
    return Se(e, this.constructor.elementStyles), e;
  }
  connectedCallback() {
    var e;
    this.renderRoot ?? (this.renderRoot = this.createRenderRoot()), this.enableUpdating(!0), (e = this._$EO) == null || e.forEach((s) => {
      var r;
      return (r = s.hostConnected) == null ? void 0 : r.call(s);
    });
  }
  enableUpdating(e) {
  }
  disconnectedCallback() {
    var e;
    (e = this._$EO) == null || e.forEach((s) => {
      var r;
      return (r = s.hostDisconnected) == null ? void 0 : r.call(s);
    });
  }
  attributeChangedCallback(e, s, r) {
    this._$AK(e, r);
  }
  _$ET(e, s) {
    var o;
    const r = this.constructor.elementProperties.get(e), i = this.constructor._$Eu(e, r);
    if (i !== void 0 && r.reflect === !0) {
      const a = (((o = r.converter) == null ? void 0 : o.toAttribute) !== void 0 ? r.converter : Q).toAttribute(s, r.type);
      this._$Em = e, a == null ? this.removeAttribute(i) : this.setAttribute(i, a), this._$Em = null;
    }
  }
  _$AK(e, s) {
    var o, a;
    const r = this.constructor, i = r._$Eh.get(e);
    if (i !== void 0 && this._$Em !== i) {
      const d = r.getPropertyOptions(i), n = typeof d.converter == "function" ? { fromAttribute: d.converter } : ((o = d.converter) == null ? void 0 : o.fromAttribute) !== void 0 ? d.converter : Q;
      this._$Em = i;
      const u = n.fromAttribute(s, d.type);
      this[i] = u ?? ((a = this._$Ej) == null ? void 0 : a.get(i)) ?? u, this._$Em = null;
    }
  }
  requestUpdate(e, s, r, i = !1, o) {
    var a;
    if (e !== void 0) {
      const d = this.constructor;
      if (i === !1 && (o = this[e]), r ?? (r = d.getPropertyOptions(e)), !((r.hasChanged ?? ae)(o, s) || r.useDefault && r.reflect && o === ((a = this._$Ej) == null ? void 0 : a.get(e)) && !this.hasAttribute(d._$Eu(e, r)))) return;
      this.C(e, s, r);
    }
    this.isUpdatePending === !1 && (this._$ES = this._$EP());
  }
  C(e, s, { useDefault: r, reflect: i, wrapped: o }, a) {
    r && !(this._$Ej ?? (this._$Ej = /* @__PURE__ */ new Map())).has(e) && (this._$Ej.set(e, a ?? s ?? this[e]), o !== !0 || a !== void 0) || (this._$AL.has(e) || (this.hasUpdated || r || (s = void 0), this._$AL.set(e, s)), i === !0 && this._$Em !== e && (this._$Eq ?? (this._$Eq = /* @__PURE__ */ new Set())).add(e));
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
    var r;
    if (!this.isUpdatePending) return;
    if (!this.hasUpdated) {
      if (this.renderRoot ?? (this.renderRoot = this.createRenderRoot()), this._$Ep) {
        for (const [o, a] of this._$Ep) this[o] = a;
        this._$Ep = void 0;
      }
      const i = this.constructor.elementProperties;
      if (i.size > 0) for (const [o, a] of i) {
        const { wrapped: d } = a, n = this[o];
        d !== !0 || this._$AL.has(o) || n === void 0 || this.C(o, void 0, a, n);
      }
    }
    let e = !1;
    const s = this._$AL;
    try {
      e = this.shouldUpdate(s), e ? (this.willUpdate(s), (r = this._$EO) == null || r.forEach((i) => {
        var o;
        return (o = i.hostUpdate) == null ? void 0 : o.call(i);
      }), this.update(s)) : this._$EM();
    } catch (i) {
      throw e = !1, this._$EM(), i;
    }
    e && this._$AE(s);
  }
  willUpdate(e) {
  }
  _$AE(e) {
    var s;
    (s = this._$EO) == null || s.forEach((r) => {
      var i;
      return (i = r.hostUpdated) == null ? void 0 : i.call(r);
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
R.elementStyles = [], R.shadowRootOptions = { mode: "open" }, R[D("elementProperties")] = /* @__PURE__ */ new Map(), R[D("finalized")] = /* @__PURE__ */ new Map(), Y == null || Y({ ReactiveElement: R }), (A.reactiveElementVersions ?? (A.reactiveElementVersions = [])).push("2.1.2");
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const I = globalThis, he = (t) => t, Z = I.trustedTypes, pe = Z ? Z.createPolicy("lit-html", { createHTML: (t) => t }) : void 0, $e = "$lit$", w = `lit$${Math.random().toFixed(9).slice(2)}$`, _e = "?" + w, ze = `<${_e}>`, U = document, j = () => U.createComment(""), L = (t) => t === null || typeof t != "object" && typeof t != "function", oe = Array.isArray, Te = (t) => oe(t) || typeof (t == null ? void 0 : t[Symbol.iterator]) == "function", ee = `[ 	
\f\r]`, H = /<(?:(!--|\/[^a-zA-Z])|(\/?[a-zA-Z][^>\s]*)|(\/?$))/g, ue = /-->/g, fe = />/g, C = RegExp(`>|${ee}(?:([^\\s"'>=/]+)(${ee}*=${ee}*(?:[^ 	
\f\r"'\`<>=]|("|')|))|$)`, "g"), me = /'/g, ge = /"/g, we = /^(?:script|style|textarea|title)$/i, Me = (t) => (e, ...s) => ({ _$litType$: t, strings: e, values: s }), l = Me(1), z = Symbol.for("lit-noChange"), c = Symbol.for("lit-nothing"), be = /* @__PURE__ */ new WeakMap(), k = U.createTreeWalker(U, 129);
function Ae(t, e) {
  if (!oe(t) || !t.hasOwnProperty("raw")) throw Error("invalid template strings array");
  return pe !== void 0 ? pe.createHTML(e) : e;
}
const Ne = (t, e) => {
  const s = t.length - 1, r = [];
  let i, o = e === 2 ? "<svg>" : e === 3 ? "<math>" : "", a = H;
  for (let d = 0; d < s; d++) {
    const n = t[d];
    let u, f, h = -1, v = 0;
    for (; v < n.length && (a.lastIndex = v, f = a.exec(n), f !== null); ) v = a.lastIndex, a === H ? f[1] === "!--" ? a = ue : f[1] !== void 0 ? a = fe : f[2] !== void 0 ? (we.test(f[2]) && (i = RegExp("</" + f[2], "g")), a = C) : f[3] !== void 0 && (a = C) : a === C ? f[0] === ">" ? (a = i ?? H, h = -1) : f[1] === void 0 ? h = -2 : (h = a.lastIndex - f[2].length, u = f[1], a = f[3] === void 0 ? C : f[3] === '"' ? ge : me) : a === ge || a === me ? a = C : a === ue || a === fe ? a = H : (a = C, i = void 0);
    const _ = a === C && t[d + 1].startsWith("/>") ? " " : "";
    o += a === H ? n + ze : h >= 0 ? (r.push(u), n.slice(0, h) + $e + n.slice(h) + w + _) : n + w + (h === -2 ? d : _);
  }
  return [Ae(t, o + (t[s] || "<?>") + (e === 2 ? "</svg>" : e === 3 ? "</math>" : "")), r];
};
class q {
  constructor({ strings: e, _$litType$: s }, r) {
    let i;
    this.parts = [];
    let o = 0, a = 0;
    const d = e.length - 1, n = this.parts, [u, f] = Ne(e, s);
    if (this.el = q.createElement(u, r), k.currentNode = this.el.content, s === 2 || s === 3) {
      const h = this.el.content.firstChild;
      h.replaceWith(...h.childNodes);
    }
    for (; (i = k.nextNode()) !== null && n.length < d; ) {
      if (i.nodeType === 1) {
        if (i.hasAttributes()) for (const h of i.getAttributeNames()) if (h.endsWith($e)) {
          const v = f[a++], _ = i.getAttribute(h).split(w), K = /([.?@])?(.*)/.exec(v);
          n.push({ type: 1, index: o, name: K[2], strings: _, ctor: K[1] === "." ? De : K[1] === "?" ? Ie : K[1] === "@" ? je : G }), i.removeAttribute(h);
        } else h.startsWith(w) && (n.push({ type: 6, index: o }), i.removeAttribute(h));
        if (we.test(i.tagName)) {
          const h = i.textContent.split(w), v = h.length - 1;
          if (v > 0) {
            i.textContent = Z ? Z.emptyScript : "";
            for (let _ = 0; _ < v; _++) i.append(h[_], j()), k.nextNode(), n.push({ type: 2, index: ++o });
            i.append(h[v], j());
          }
        }
      } else if (i.nodeType === 8) if (i.data === _e) n.push({ type: 2, index: o });
      else {
        let h = -1;
        for (; (h = i.data.indexOf(w, h + 1)) !== -1; ) n.push({ type: 7, index: o }), h += w.length - 1;
      }
      o++;
    }
  }
  static createElement(e, s) {
    const r = U.createElement("template");
    return r.innerHTML = e, r;
  }
}
function T(t, e, s = t, r) {
  var a, d;
  if (e === z) return e;
  let i = r !== void 0 ? (a = s._$Co) == null ? void 0 : a[r] : s._$Cl;
  const o = L(e) ? void 0 : e._$litDirective$;
  return (i == null ? void 0 : i.constructor) !== o && ((d = i == null ? void 0 : i._$AO) == null || d.call(i, !1), o === void 0 ? i = void 0 : (i = new o(t), i._$AT(t, s, r)), r !== void 0 ? (s._$Co ?? (s._$Co = []))[r] = i : s._$Cl = i), i !== void 0 && (e = T(t, i._$AS(t, e.values), i, r)), e;
}
class He {
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
    const { el: { content: s }, parts: r } = this._$AD, i = ((e == null ? void 0 : e.creationScope) ?? U).importNode(s, !0);
    k.currentNode = i;
    let o = k.nextNode(), a = 0, d = 0, n = r[0];
    for (; n !== void 0; ) {
      if (a === n.index) {
        let u;
        n.type === 2 ? u = new V(o, o.nextSibling, this, e) : n.type === 1 ? u = new n.ctor(o, n.name, n.strings, this, e) : n.type === 6 && (u = new Le(o, this, e)), this._$AV.push(u), n = r[++d];
      }
      a !== (n == null ? void 0 : n.index) && (o = k.nextNode(), a++);
    }
    return k.currentNode = U, i;
  }
  p(e) {
    let s = 0;
    for (const r of this._$AV) r !== void 0 && (r.strings !== void 0 ? (r._$AI(e, r, s), s += r.strings.length - 2) : r._$AI(e[s])), s++;
  }
}
class V {
  get _$AU() {
    var e;
    return ((e = this._$AM) == null ? void 0 : e._$AU) ?? this._$Cv;
  }
  constructor(e, s, r, i) {
    this.type = 2, this._$AH = c, this._$AN = void 0, this._$AA = e, this._$AB = s, this._$AM = r, this.options = i, this._$Cv = (i == null ? void 0 : i.isConnected) ?? !0;
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
    e = T(this, e, s), L(e) ? e === c || e == null || e === "" ? (this._$AH !== c && this._$AR(), this._$AH = c) : e !== this._$AH && e !== z && this._(e) : e._$litType$ !== void 0 ? this.$(e) : e.nodeType !== void 0 ? this.T(e) : Te(e) ? this.k(e) : this._(e);
  }
  O(e) {
    return this._$AA.parentNode.insertBefore(e, this._$AB);
  }
  T(e) {
    this._$AH !== e && (this._$AR(), this._$AH = this.O(e));
  }
  _(e) {
    this._$AH !== c && L(this._$AH) ? this._$AA.nextSibling.data = e : this.T(U.createTextNode(e)), this._$AH = e;
  }
  $(e) {
    var o;
    const { values: s, _$litType$: r } = e, i = typeof r == "number" ? this._$AC(e) : (r.el === void 0 && (r.el = q.createElement(Ae(r.h, r.h[0]), this.options)), r);
    if (((o = this._$AH) == null ? void 0 : o._$AD) === i) this._$AH.p(s);
    else {
      const a = new He(i, this), d = a.u(this.options);
      a.p(s), this.T(d), this._$AH = a;
    }
  }
  _$AC(e) {
    let s = be.get(e.strings);
    return s === void 0 && be.set(e.strings, s = new q(e)), s;
  }
  k(e) {
    oe(this._$AH) || (this._$AH = [], this._$AR());
    const s = this._$AH;
    let r, i = 0;
    for (const o of e) i === s.length ? s.push(r = new V(this.O(j()), this.O(j()), this, this.options)) : r = s[i], r._$AI(o), i++;
    i < s.length && (this._$AR(r && r._$AB.nextSibling, i), s.length = i);
  }
  _$AR(e = this._$AA.nextSibling, s) {
    var r;
    for ((r = this._$AP) == null ? void 0 : r.call(this, !1, !0, s); e !== this._$AB; ) {
      const i = he(e).nextSibling;
      he(e).remove(), e = i;
    }
  }
  setConnected(e) {
    var s;
    this._$AM === void 0 && (this._$Cv = e, (s = this._$AP) == null || s.call(this, e));
  }
}
class G {
  get tagName() {
    return this.element.tagName;
  }
  get _$AU() {
    return this._$AM._$AU;
  }
  constructor(e, s, r, i, o) {
    this.type = 1, this._$AH = c, this._$AN = void 0, this.element = e, this.name = s, this._$AM = i, this.options = o, r.length > 2 || r[0] !== "" || r[1] !== "" ? (this._$AH = Array(r.length - 1).fill(new String()), this.strings = r) : this._$AH = c;
  }
  _$AI(e, s = this, r, i) {
    const o = this.strings;
    let a = !1;
    if (o === void 0) e = T(this, e, s, 0), a = !L(e) || e !== this._$AH && e !== z, a && (this._$AH = e);
    else {
      const d = e;
      let n, u;
      for (e = o[0], n = 0; n < o.length - 1; n++) u = T(this, d[r + n], s, n), u === z && (u = this._$AH[n]), a || (a = !L(u) || u !== this._$AH[n]), u === c ? e = c : e !== c && (e += (u ?? "") + o[n + 1]), this._$AH[n] = u;
    }
    a && !i && this.j(e);
  }
  j(e) {
    e === c ? this.element.removeAttribute(this.name) : this.element.setAttribute(this.name, e ?? "");
  }
}
class De extends G {
  constructor() {
    super(...arguments), this.type = 3;
  }
  j(e) {
    this.element[this.name] = e === c ? void 0 : e;
  }
}
class Ie extends G {
  constructor() {
    super(...arguments), this.type = 4;
  }
  j(e) {
    this.element.toggleAttribute(this.name, !!e && e !== c);
  }
}
class je extends G {
  constructor(e, s, r, i, o) {
    super(e, s, r, i, o), this.type = 5;
  }
  _$AI(e, s = this) {
    if ((e = T(this, e, s, 0) ?? c) === z) return;
    const r = this._$AH, i = e === c && r !== c || e.capture !== r.capture || e.once !== r.once || e.passive !== r.passive, o = e !== c && (r === c || i);
    i && this.element.removeEventListener(this.name, this, r), o && this.element.addEventListener(this.name, this, e), this._$AH = e;
  }
  handleEvent(e) {
    var s;
    typeof this._$AH == "function" ? this._$AH.call(((s = this.options) == null ? void 0 : s.host) ?? this.element, e) : this._$AH.handleEvent(e);
  }
}
class Le {
  constructor(e, s, r) {
    this.element = e, this.type = 6, this._$AN = void 0, this._$AM = s, this.options = r;
  }
  get _$AU() {
    return this._$AM._$AU;
  }
  _$AI(e) {
    T(this, e);
  }
}
const te = I.litHtmlPolyfillSupport;
te == null || te(q, V), (I.litHtmlVersions ?? (I.litHtmlVersions = [])).push("3.3.2");
const qe = (t, e, s) => {
  const r = (s == null ? void 0 : s.renderBefore) ?? e;
  let i = r._$litPart$;
  if (i === void 0) {
    const o = (s == null ? void 0 : s.renderBefore) ?? null;
    r._$litPart$ = i = new V(e.insertBefore(j(), o), o, void 0, s ?? {});
  }
  return i._$AI(t), i;
};
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const P = globalThis;
class y extends R {
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
    this.hasUpdated || (this.renderOptions.isConnected = this.isConnected), super.update(e), this._$Do = qe(s, this.renderRoot, this.renderOptions);
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
    return z;
  }
}
var ve;
y._$litElement$ = !0, y.finalized = !0, (ve = P.litElementHydrateSupport) == null || ve.call(P, { LitElement: y });
const se = P.litElementPolyfillSupport;
se == null || se({ LitElement: y });
(P.litElementVersions ?? (P.litElementVersions = [])).push("4.2.2");
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const B = (t) => (e, s) => {
  s !== void 0 ? s.addInitializer(() => {
    customElements.define(t, e);
  }) : customElements.define(t, e);
};
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
const We = { attribute: !0, type: String, converter: Q, reflect: !1, hasChanged: ae }, Ve = (t = We, e, s) => {
  const { kind: r, metadata: i } = s;
  let o = globalThis.litPropertyMetadata.get(i);
  if (o === void 0 && globalThis.litPropertyMetadata.set(i, o = /* @__PURE__ */ new Map()), r === "setter" && ((t = Object.create(t)).wrapped = !0), o.set(s.name, t), r === "accessor") {
    const { name: a } = s;
    return { set(d) {
      const n = e.get.call(this);
      e.set.call(this, d), this.requestUpdate(a, n, t, !0, d);
    }, init(d) {
      return d !== void 0 && this.C(a, void 0, t, d), d;
    } };
  }
  if (r === "setter") {
    const { name: a } = s;
    return function(d) {
      const n = this[a];
      e.call(this, d), this.requestUpdate(a, n, t, !0, d);
    };
  }
  throw Error("Unsupported decorator location: " + r);
};
function $(t) {
  return (e, s) => typeof s == "object" ? Ve(t, e, s) : ((r, i, o) => {
    const a = i.hasOwnProperty(o);
    return i.constructor.createProperty(o, r), a ? Object.getOwnPropertyDescriptor(i, o) : void 0;
  })(t, e, s);
}
/**
 * @license
 * Copyright 2017 Google LLC
 * SPDX-License-Identifier: BSD-3-Clause
 */
function p(t) {
  return $({ ...t, state: !0, attribute: !1 });
}
function Be(t, e) {
  const s = new WebSocket(t);
  return s.onmessage = (r) => {
    var i, o, a, d;
    try {
      const n = JSON.parse(r.data);
      ((o = (i = n.type) == null ? void 0 : i.startsWith) != null && o.call(i, "scm.") || (d = (a = n.channel) == null ? void 0 : a.startsWith) != null && d.call(a, "scm.")) && e(n);
    } catch {
    }
  }, s;
}
class X {
  constructor(e = "") {
    this.baseUrl = e;
  }
  get base() {
    return `${this.baseUrl}/api/v1/scm`;
  }
  async request(e, s) {
    var o, a;
    const r = await fetch(`${this.base}${e}`, s), i = await r.json().catch(() => null);
    if (!r.ok)
      throw new Error(((o = i == null ? void 0 : i.error) == null ? void 0 : o.message) ?? `Request failed (${r.status})`);
    if (!(i != null && i.success))
      throw new Error(((a = i == null ? void 0 : i.error) == null ? void 0 : a.message) ?? "Request failed");
    return i.data;
  }
  marketplace(e, s) {
    const r = new URLSearchParams();
    e && r.set("q", e), s && r.set("category", s);
    const i = r.toString();
    return this.request(`/marketplace${i ? `?${i}` : ""}`);
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
var Fe = Object.defineProperty, Ke = Object.getOwnPropertyDescriptor, b = (t, e, s, r) => {
  for (var i = r > 1 ? void 0 : r ? Ke(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (i = (r ? a(e, s, i) : a(i)) || i);
  return r && i && Fe(e, s, i), i;
};
let m = class extends y {
  constructor() {
    super(...arguments), this.apiUrl = "", this.category = "", this.modules = [], this.categories = [], this.searchQuery = "", this.activeCategory = "", this.loading = !0, this.error = "", this.installing = /* @__PURE__ */ new Set();
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new X(this.apiUrl), this.activeCategory = this.category, this.loadModules();
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
      <div class="toolbar">
        <input
          type="text"
          class="search"
          placeholder="Search providers\u2026"
          .value=${this.searchQuery}
          @input=${this.handleSearch}
        />
      </div>

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
      ${this.error ? l`<div class="error">${this.error}</div>` : c}
      ${this.loading ? l`<div class="loading">Loading marketplace\u2026</div>` : this.modules.length === 0 ? l`<div class="empty">No providers found.</div>` : l`
              <div class="grid">
                ${this.modules.map(
      (t) => l`
                    <div class="card">
                      <div class="card-header">
                        <div>
                          <div class="card-name">${t.name}</div>
                          <div class="card-code">${t.code}</div>
                        </div>
                        ${t.category ? l`<span class="card-category">${t.category}</span>` : c}
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
    `;
  }
};
m.styles = W`
    :host {
      display: block;
      font-family: system-ui, -apple-system, sans-serif;
    }

    .toolbar {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-bottom: 1rem;
    }

    .search {
      flex: 1;
      padding: 0.5rem 0.75rem;
      border: 1px solid #d1d5db;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      outline: none;
    }

    .search:focus {
      border-colour: #6366f1;
      box-shadow: 0 0 0 2px rgba(99, 102, 241, 0.2);
    }

    .categories {
      display: flex;
      gap: 0.25rem;
      flex-wrap: wrap;
    }

    .category-btn {
      padding: 0.25rem 0.75rem;
      border: 1px solid #e5e7eb;
      border-radius: 1rem;
      background: #fff;
      font-size: 0.75rem;
      cursor: pointer;
      transition: all 0.15s;
    }

    .category-btn:hover {
      background: #f3f4f6;
    }

    .category-btn.active {
      background: #6366f1;
      colour: #fff;
      border-colour: #6366f1;
    }

    .grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
      gap: 1rem;
      margin-top: 1rem;
    }

    .card {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1rem;
      background: #fff;
      transition: box-shadow 0.15s;
    }

    .card:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
    }

    .card-header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 0.5rem;
    }

    .card-name {
      font-weight: 600;
      font-size: 0.9375rem;
    }

    .card-code {
      font-size: 0.75rem;
      colour: #6b7280;
      font-family: monospace;
    }

    .card-category {
      font-size: 0.6875rem;
      padding: 0.125rem 0.5rem;
      background: #f3f4f6;
      border-radius: 1rem;
      colour: #6b7280;
    }

    .card-actions {
      margin-top: 0.75rem;
      display: flex;
      gap: 0.5rem;
    }

    button.install {
      padding: 0.375rem 1rem;
      background: #6366f1;
      colour: #fff;
      border: none;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: background 0.15s;
    }

    button.install:hover {
      background: #4f46e5;
    }

    button.install:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      padding: 0.375rem 1rem;
      background: #fff;
      colour: #dc2626;
      border: 1px solid #dc2626;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
    }

    button.remove:hover {
      background: #fef2f2;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      colour: #9ca3af;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      colour: #6b7280;
    }

    .error {
      colour: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border-radius: 0.375rem;
      font-size: 0.875rem;
    }
  `;
b([
  $({ attribute: "api-url" })
], m.prototype, "apiUrl", 2);
b([
  $()
], m.prototype, "category", 2);
b([
  p()
], m.prototype, "modules", 2);
b([
  p()
], m.prototype, "categories", 2);
b([
  p()
], m.prototype, "searchQuery", 2);
b([
  p()
], m.prototype, "activeCategory", 2);
b([
  p()
], m.prototype, "loading", 2);
b([
  p()
], m.prototype, "error", 2);
b([
  p()
], m.prototype, "installing", 2);
m = b([
  B("core-scm-marketplace")
], m);
var Je = Object.defineProperty, Qe = Object.getOwnPropertyDescriptor, M = (t, e, s, r) => {
  for (var i = r > 1 ? void 0 : r ? Qe(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (i = (r ? a(e, s, i) : a(i)) || i);
  return r && i && Je(e, s, i), i;
};
let x = class extends y {
  constructor() {
    super(...arguments), this.apiUrl = "", this.modules = [], this.loading = !0, this.error = "", this.updating = /* @__PURE__ */ new Set();
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new X(this.apiUrl), this.loadInstalled();
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
    `;
  }
};
x.styles = W`
    :host {
      display: block;
      font-family: system-ui, -apple-system, sans-serif;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.5rem;
    }

    .item {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1rem;
      background: #fff;
      display: flex;
      justify-content: space-between;
      align-items: center;
      transition: box-shadow 0.15s;
    }

    .item:hover {
      box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
    }

    .item-info {
      flex: 1;
    }

    .item-name {
      font-weight: 600;
      font-size: 0.9375rem;
    }

    .item-meta {
      font-size: 0.75rem;
      colour: #6b7280;
      margin-top: 0.25rem;
      display: flex;
      gap: 1rem;
    }

    .item-code {
      font-family: monospace;
    }

    .item-actions {
      display: flex;
      gap: 0.5rem;
    }

    button {
      padding: 0.375rem 0.75rem;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: background 0.15s;
    }

    button.update {
      background: #fff;
      colour: #6366f1;
      border: 1px solid #6366f1;
    }

    button.update:hover {
      background: #eef2ff;
    }

    button.update:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }

    button.remove {
      background: #fff;
      colour: #dc2626;
      border: 1px solid #dc2626;
    }

    button.remove:hover {
      background: #fef2f2;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      colour: #9ca3af;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      colour: #6b7280;
    }

    .error {
      colour: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }
  `;
M([
  $({ attribute: "api-url" })
], x.prototype, "apiUrl", 2);
M([
  p()
], x.prototype, "modules", 2);
M([
  p()
], x.prototype, "loading", 2);
M([
  p()
], x.prototype, "error", 2);
M([
  p()
], x.prototype, "updating", 2);
x = M([
  B("core-scm-installed")
], x);
var Ze = Object.defineProperty, Ge = Object.getOwnPropertyDescriptor, E = (t, e, s, r) => {
  for (var i = r > 1 ? void 0 : r ? Ge(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (i = (r ? a(e, s, i) : a(i)) || i);
  return r && i && Ze(e, s, i), i;
};
let g = class extends y {
  constructor() {
    super(...arguments), this.apiUrl = "", this.path = "", this.manifest = null, this.loading = !0, this.error = "", this.verifyKey = "", this.verifyResult = null;
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new X(this.apiUrl), this.loadManifest();
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
                ${s.items.map((r) => l`<div class="perm-item">${r}</div>`)}
              </div>
            `
    )}
        </div>
      </div>
    `;
  }
  render() {
    if (this.loading)
      return l`<div class="loading">Loading manifest\u2026</div>`;
    if (this.error && !this.manifest)
      return l`<div class="error">${this.error}</div>`;
    if (!this.manifest)
      return l`<div class="empty">No manifest found. Create a .core/manifest.yaml to get started.</div>`;
    const t = this.manifest, e = !!t.sign;
    return l`
      ${this.error ? l`<div class="error">${this.error}</div>` : c}
      <div class="manifest">
        <div class="header">
          <div>
            <h3>${t.name}</h3>
            <span class="code">${t.code}</span>
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
      ([s, r]) => l`
                      <span class="slot-key">${s}</span>
                      <span class="slot-value">${r}</span>
                    `
    )}
                </div>
              </div>
            ` : c}

        ${this.renderPermissions(t.permissions)}
        ${t.modules && t.modules.length > 0 ? l`
              <div class="field">
                <div class="field-label">Modules</div>
                ${t.modules.map((s) => l`<div class="code" style="margin-bottom:0.25rem">${s}</div>`)}
              </div>
            ` : c}

        <div class="signature ${e ? this.verifyResult ? this.verifyResult.valid ? "signed" : "invalid" : "signed" : "unsigned"}">
          <span class="badge ${e && this.verifyResult ? this.verifyResult.valid ? "verified" : "invalid" : "unsigned"}">
            ${e ? this.verifyResult ? this.verifyResult.valid ? "Verified" : "Invalid" : "Signed" : "Unsigned"}
          </span>
          ${e ? l`<span>Signature present</span>` : l`<span>No signature</span>`}
        </div>

        <div class="actions">
          <input
            type="text"
            class="verify-input"
            placeholder="Public key (hex)\u2026"
            .value=${this.verifyKey}
            @input=${(s) => this.verifyKey = s.target.value}
          />
          <button @click=${this.handleVerify}>Verify</button>
          <button class="primary" @click=${this.handleSign}>Sign</button>
        </div>
      </div>
    `;
  }
};
g.styles = W`
    :host {
      display: block;
      font-family: system-ui, -apple-system, sans-serif;
    }

    .manifest {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 1.25rem;
      background: #fff;
    }

    .header {
      display: flex;
      justify-content: space-between;
      align-items: flex-start;
      margin-bottom: 1rem;
    }

    h3 {
      margin: 0;
      font-size: 1.125rem;
      font-weight: 600;
    }

    .version {
      font-size: 0.75rem;
      padding: 0.125rem 0.5rem;
      background: #f3f4f6;
      border-radius: 1rem;
      colour: #6b7280;
    }

    .field {
      margin-bottom: 0.75rem;
    }

    .field-label {
      font-size: 0.75rem;
      font-weight: 600;
      colour: #6b7280;
      text-transform: uppercase;
      letter-spacing: 0.025em;
      margin-bottom: 0.25rem;
    }

    .field-value {
      font-size: 0.875rem;
    }

    .code {
      font-family: monospace;
      font-size: 0.8125rem;
      background: #f9fafb;
      padding: 0.25rem 0.5rem;
      border-radius: 0.25rem;
    }

    .slots {
      display: grid;
      grid-template-columns: auto 1fr;
      gap: 0.25rem 1rem;
      font-size: 0.8125rem;
    }

    .slot-key {
      font-weight: 600;
      colour: #374151;
    }

    .slot-value {
      font-family: monospace;
      colour: #6b7280;
    }

    .permissions-grid {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
      gap: 0.5rem;
    }

    .perm-group {
      border: 1px solid #e5e7eb;
      border-radius: 0.375rem;
      padding: 0.5rem;
    }

    .perm-group-label {
      font-size: 0.6875rem;
      font-weight: 700;
      colour: #6b7280;
      text-transform: uppercase;
      margin-bottom: 0.25rem;
    }

    .perm-item {
      font-size: 0.8125rem;
      font-family: monospace;
      colour: #374151;
    }

    .signature {
      display: flex;
      align-items: center;
      gap: 0.75rem;
      margin-top: 1rem;
      padding: 0.75rem;
      border-radius: 0.375rem;
      font-size: 0.875rem;
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
      font-weight: 600;
      padding: 0.125rem 0.5rem;
      border-radius: 1rem;
    }

    .badge.verified {
      background: #dcfce7;
      colour: #166534;
    }

    .badge.unsigned {
      background: #fef3c7;
      colour: #92400e;
    }

    .badge.invalid {
      background: #fee2e2;
      colour: #991b1b;
    }

    .actions {
      margin-top: 1rem;
      display: flex;
      gap: 0.5rem;
    }

    .verify-input {
      flex: 1;
      padding: 0.375rem 0.75rem;
      border: 1px solid #d1d5db;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      font-family: monospace;
    }

    button {
      padding: 0.375rem 1rem;
      border-radius: 0.375rem;
      font-size: 0.8125rem;
      cursor: pointer;
      border: 1px solid #d1d5db;
      background: #fff;
      transition: background 0.15s;
    }

    button:hover {
      background: #f3f4f6;
    }

    button.primary {
      background: #6366f1;
      colour: #fff;
      border-colour: #6366f1;
    }

    button.primary:hover {
      background: #4f46e5;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      colour: #9ca3af;
      font-size: 0.875rem;
    }

    .error {
      colour: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border-radius: 0.375rem;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      colour: #6b7280;
    }
  `;
E([
  $({ attribute: "api-url" })
], g.prototype, "apiUrl", 2);
E([
  $()
], g.prototype, "path", 2);
E([
  p()
], g.prototype, "manifest", 2);
E([
  p()
], g.prototype, "loading", 2);
E([
  p()
], g.prototype, "error", 2);
E([
  p()
], g.prototype, "verifyKey", 2);
E([
  p()
], g.prototype, "verifyResult", 2);
g = E([
  B("core-scm-manifest")
], g);
var Xe = Object.defineProperty, Ye = Object.getOwnPropertyDescriptor, F = (t, e, s, r) => {
  for (var i = r > 1 ? void 0 : r ? Ye(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (i = (r ? a(e, s, i) : a(i)) || i);
  return r && i && Xe(e, s, i), i;
};
let O = class extends y {
  constructor() {
    super(...arguments), this.apiUrl = "", this.repos = [], this.loading = !0, this.error = "";
  }
  connectedCallback() {
    super.connectedCallback(), this.api = new X(this.apiUrl), this.loadRegistry();
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
    `;
  }
};
O.styles = W`
    :host {
      display: block;
      font-family: system-ui, -apple-system, sans-serif;
    }

    .list {
      display: flex;
      flex-direction: column;
      gap: 0.375rem;
    }

    .repo {
      border: 1px solid #e5e7eb;
      border-radius: 0.5rem;
      padding: 0.75rem 1rem;
      background: #fff;
      display: flex;
      justify-content: space-between;
      align-items: center;
    }

    .repo-info {
      flex: 1;
    }

    .repo-name {
      font-weight: 600;
      font-size: 0.9375rem;
      font-family: monospace;
    }

    .repo-desc {
      font-size: 0.8125rem;
      colour: #6b7280;
      margin-top: 0.125rem;
    }

    .repo-meta {
      display: flex;
      gap: 0.5rem;
      align-items: center;
      margin-top: 0.25rem;
    }

    .type-badge {
      font-size: 0.6875rem;
      padding: 0.0625rem 0.5rem;
      border-radius: 1rem;
      font-weight: 600;
    }

    .type-badge.foundation {
      background: #dbeafe;
      colour: #1e40af;
    }

    .type-badge.module {
      background: #f3e8ff;
      colour: #6b21a8;
    }

    .type-badge.product {
      background: #dcfce7;
      colour: #166534;
    }

    .type-badge.template {
      background: #fef3c7;
      colour: #92400e;
    }

    .deps {
      font-size: 0.75rem;
      colour: #9ca3af;
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
    }

    .status-dot.missing {
      background: #ef4444;
    }

    .status-label {
      font-size: 0.75rem;
      colour: #6b7280;
    }

    .empty {
      text-align: center;
      padding: 2rem;
      colour: #9ca3af;
      font-size: 0.875rem;
    }

    .loading {
      text-align: center;
      padding: 2rem;
      colour: #6b7280;
    }

    .error {
      colour: #dc2626;
      padding: 0.75rem;
      background: #fef2f2;
      border-radius: 0.375rem;
      font-size: 0.875rem;
      margin-bottom: 1rem;
    }
  `;
F([
  $({ attribute: "api-url" })
], O.prototype, "apiUrl", 2);
F([
  p()
], O.prototype, "repos", 2);
F([
  p()
], O.prototype, "loading", 2);
F([
  p()
], O.prototype, "error", 2);
O = F([
  B("core-scm-registry")
], O);
var et = Object.defineProperty, tt = Object.getOwnPropertyDescriptor, N = (t, e, s, r) => {
  for (var i = r > 1 ? void 0 : r ? tt(e, s) : e, o = t.length - 1, a; o >= 0; o--)
    (a = t[o]) && (i = (r ? a(e, s, i) : a(i)) || i);
  return r && i && et(e, s, i), i;
};
let S = class extends y {
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
    this.disconnectWs(), this.wsUrl && (this.ws = Be(this.wsUrl, (t) => {
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
        <span class="title">SCM</span>
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
S.styles = W`
    :host {
      display: flex;
      flex-direction: column;
      font-family: system-ui, -apple-system, sans-serif;
      height: 100%;
      background: #fafafa;
    }

    /* H — Header */
    .header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.75rem 1rem;
      background: #fff;
      border-bottom: 1px solid #e5e7eb;
    }

    .title {
      font-weight: 700;
      font-size: 1rem;
      colour: #111827;
    }

    .refresh-btn {
      padding: 0.375rem 0.75rem;
      border: 1px solid #d1d5db;
      border-radius: 0.375rem;
      background: #fff;
      font-size: 0.8125rem;
      cursor: pointer;
      transition: background 0.15s;
    }

    .refresh-btn:hover {
      background: #f3f4f6;
    }

    /* H-L — Tabs */
    .tabs {
      display: flex;
      gap: 0;
      background: #fff;
      border-bottom: 1px solid #e5e7eb;
      padding: 0 1rem;
    }

    .tab {
      padding: 0.625rem 1rem;
      font-size: 0.8125rem;
      font-weight: 500;
      colour: #6b7280;
      cursor: pointer;
      border-bottom: 2px solid transparent;
      transition: all 0.15s;
      background: none;
      border-top: none;
      border-left: none;
      border-right: none;
    }

    .tab:hover {
      colour: #374151;
    }

    .tab.active {
      colour: #6366f1;
      border-bottom-colour: #6366f1;
    }

    /* C — Content */
    .content {
      flex: 1;
      padding: 1rem;
      overflow-y: auto;
    }

    /* F — Footer / Status bar */
    .footer {
      display: flex;
      justify-content: space-between;
      align-items: center;
      padding: 0.5rem 1rem;
      background: #fff;
      border-top: 1px solid #e5e7eb;
      font-size: 0.75rem;
      colour: #9ca3af;
    }

    .ws-status {
      display: flex;
      align-items: center;
      gap: 0.375rem;
    }

    .ws-dot {
      width: 0.5rem;
      height: 0.5rem;
      border-radius: 50%;
    }

    .ws-dot.connected {
      background: #22c55e;
    }

    .ws-dot.disconnected {
      background: #ef4444;
    }

    .ws-dot.idle {
      background: #d1d5db;
    }
  `;
N([
  $({ attribute: "api-url" })
], S.prototype, "apiUrl", 2);
N([
  $({ attribute: "ws-url" })
], S.prototype, "wsUrl", 2);
N([
  p()
], S.prototype, "activeTab", 2);
N([
  p()
], S.prototype, "wsConnected", 2);
N([
  p()
], S.prototype, "lastEvent", 2);
S = N([
  B("core-scm-panel")
], S);
export {
  X as ScmApi,
  x as ScmInstalled,
  g as ScmManifest,
  m as ScmMarketplace,
  S as ScmPanel,
  O as ScmRegistry,
  Be as connectScmEvents
};
