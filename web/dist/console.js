var CoreConsole = (() => {
  var __defProp = Object.defineProperty;
  var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
  var __getOwnPropNames = Object.getOwnPropertyNames;
  var __hasOwnProp = Object.prototype.hasOwnProperty;
  var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
  var __export = (target, all) => {
    for (var name in all)
      __defProp(target, name, { get: all[name], enumerable: true });
  };
  var __copyProps = (to, from, except, desc) => {
    if (from && typeof from === "object" || typeof from === "function") {
      for (let key of __getOwnPropNames(from))
        if (!__hasOwnProp.call(to, key) && key !== except)
          __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
    }
    return to;
  };
  var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);
  var __publicField = (obj, key, value) => __defNormalProp(obj, typeof key !== "symbol" ? key + "" : key, value);

  // src/main.js
  var main_exports = {};
  __export(main_exports, {
    mount: () => mount
  });

  // src/core/format.js
  function esc(s) {
    return String(s == null ? "" : s).replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#39;");
  }
  function safeHref(u) {
    const s = String(u || "").trim();
    if (!s) return "";
    if (/^https?:\/\//i.test(s)) return s;
    if (/^(#|\/|\.\.?\/)/.test(s)) return s;
    return "";
  }
  function pad(n) {
    return n < 10 ? "0" + n : "" + n;
  }
  function toDate(v) {
    if (v instanceof Date) return v;
    if (typeof v === "number") return new Date(v);
    return new Date(String(v == null ? "" : v).replace(" ", "T"));
  }
  function fmtTime(v) {
    const d = toDate(v);
    return isNaN(d) ? "\u2014" : pad(d.getHours()) + ":" + pad(d.getMinutes());
  }
  function fmtTimeS(v) {
    const d = toDate(v);
    return isNaN(d) ? "\u2014" : fmtTime(d) + ":" + pad(d.getSeconds());
  }
  function fmtDate(v) {
    const d = toDate(v);
    return isNaN(d) ? "\u2014" : d.getFullYear() + "-" + pad(d.getMonth() + 1) + "-" + pad(d.getDate());
  }
  function fmtAbs(v) {
    const d = toDate(v);
    if (isNaN(d)) return "\u65F6\u95F4\u672A\u77E5";
    let zone = "";
    try {
      zone = Intl.DateTimeFormat().resolvedOptions().timeZone || "";
    } catch {
      zone = "";
    }
    return fmtDate(d) + " " + fmtTimeS(d) + (zone ? " " + zone : "");
  }
  function relAge(ts, nowMs) {
    const t2 = toDate(ts).getTime();
    if (isNaN(t2)) return "";
    const d = Math.max(0, Math.round(((nowMs || Date.now()) - t2) / 1e3));
    if (d < 60) return d + "s \u524D";
    if (d < 3600) return Math.round(d / 60) + "m \u524D";
    if (d < 86400) return Math.round(d / 3600) + "h \u524D";
    return Math.round(d / 86400) + "d \u524D";
  }
  function timeHtml(value) {
    const d = toDate(value);
    const iso = isNaN(d) ? "" : d.toISOString();
    return '<time class="mono" datetime="' + esc(iso) + '" title="' + esc(fmtAbs(d)) + '">' + esc(relAge(value) || "\u65F6\u95F4\u672A\u77E5") + "</time>";
  }
  function fmtNum(n) {
    if (n == null || isNaN(Number(n))) return "\u2014";
    return Number(n).toLocaleString("zh-CN", { maximumFractionDigits: 1 });
  }
  function trunc(s, n) {
    const t2 = String(s == null ? "" : s);
    return t2.length > n ? t2.slice(0, n) + "\u2026" : t2;
  }

  // node_modules/fast-json-patch/module/core.mjs
  var core_exports = {};
  __export(core_exports, {
    JsonPatchError: () => JsonPatchError,
    _areEquals: () => _areEquals,
    applyOperation: () => applyOperation,
    applyPatch: () => applyPatch,
    applyReducer: () => applyReducer,
    deepClone: () => deepClone,
    getValueByPointer: () => getValueByPointer,
    validate: () => validate,
    validator: () => validator
  });

  // node_modules/fast-json-patch/module/helpers.mjs
  var __extends = /* @__PURE__ */ (function() {
    var extendStatics = function(d, b) {
      extendStatics = Object.setPrototypeOf || { __proto__: [] } instanceof Array && function(d2, b2) {
        d2.__proto__ = b2;
      } || function(d2, b2) {
        for (var p in b2) if (b2.hasOwnProperty(p)) d2[p] = b2[p];
      };
      return extendStatics(d, b);
    };
    return function(d, b) {
      extendStatics(d, b);
      function __() {
        this.constructor = d;
      }
      d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
  })();
  var _hasOwnProperty = Object.prototype.hasOwnProperty;
  function hasOwnProperty(obj, key) {
    return _hasOwnProperty.call(obj, key);
  }
  function _objectKeys(obj) {
    if (Array.isArray(obj)) {
      var keys_1 = new Array(obj.length);
      for (var k = 0; k < keys_1.length; k++) {
        keys_1[k] = "" + k;
      }
      return keys_1;
    }
    if (Object.keys) {
      return Object.keys(obj);
    }
    var keys = [];
    for (var i in obj) {
      if (hasOwnProperty(obj, i)) {
        keys.push(i);
      }
    }
    return keys;
  }
  function _deepClone(obj) {
    switch (typeof obj) {
      case "object":
        return JSON.parse(JSON.stringify(obj));
      //Faster than ES5 clone - http://jsperf.com/deep-cloning-of-objects/5
      case "undefined":
        return null;
      //this is how JSON.stringify behaves for array items
      default:
        return obj;
    }
  }
  function isInteger(str2) {
    var i = 0;
    var len = str2.length;
    var charCode;
    while (i < len) {
      charCode = str2.charCodeAt(i);
      if (charCode >= 48 && charCode <= 57) {
        i++;
        continue;
      }
      return false;
    }
    return true;
  }
  function escapePathComponent(path) {
    if (path.indexOf("/") === -1 && path.indexOf("~") === -1)
      return path;
    return path.replace(/~/g, "~0").replace(/\//g, "~1");
  }
  function unescapePathComponent(path) {
    return path.replace(/~1/g, "/").replace(/~0/g, "~");
  }
  function hasUndefined(obj) {
    if (obj === void 0) {
      return true;
    }
    if (obj) {
      if (Array.isArray(obj)) {
        for (var i_1 = 0, len = obj.length; i_1 < len; i_1++) {
          if (hasUndefined(obj[i_1])) {
            return true;
          }
        }
      } else if (typeof obj === "object") {
        var objKeys = _objectKeys(obj);
        var objKeysLength = objKeys.length;
        for (var i = 0; i < objKeysLength; i++) {
          if (hasUndefined(obj[objKeys[i]])) {
            return true;
          }
        }
      }
    }
    return false;
  }
  function patchErrorMessageFormatter(message, args) {
    var messageParts = [message];
    for (var key in args) {
      var value = typeof args[key] === "object" ? JSON.stringify(args[key], null, 2) : args[key];
      if (typeof value !== "undefined") {
        messageParts.push(key + ": " + value);
      }
    }
    return messageParts.join("\n");
  }
  var PatchError = (
    /** @class */
    (function(_super) {
      __extends(PatchError2, _super);
      function PatchError2(message, name, index, operation, tree) {
        var _newTarget = this.constructor;
        var _this = _super.call(this, patchErrorMessageFormatter(message, { name, index, operation, tree })) || this;
        _this.name = name;
        _this.index = index;
        _this.operation = operation;
        _this.tree = tree;
        Object.setPrototypeOf(_this, _newTarget.prototype);
        _this.message = patchErrorMessageFormatter(message, { name, index, operation, tree });
        return _this;
      }
      return PatchError2;
    })(Error)
  );

  // node_modules/fast-json-patch/module/core.mjs
  var JsonPatchError = PatchError;
  var deepClone = _deepClone;
  var objOps = {
    add: function(obj, key, document2) {
      obj[key] = this.value;
      return { newDocument: document2 };
    },
    remove: function(obj, key, document2) {
      var removed = obj[key];
      delete obj[key];
      return { newDocument: document2, removed };
    },
    replace: function(obj, key, document2) {
      var removed = obj[key];
      obj[key] = this.value;
      return { newDocument: document2, removed };
    },
    move: function(obj, key, document2) {
      var removed = getValueByPointer(document2, this.path);
      if (removed) {
        removed = _deepClone(removed);
      }
      var originalValue = applyOperation(document2, { op: "remove", path: this.from }).removed;
      applyOperation(document2, { op: "add", path: this.path, value: originalValue });
      return { newDocument: document2, removed };
    },
    copy: function(obj, key, document2) {
      var valueToCopy = getValueByPointer(document2, this.from);
      applyOperation(document2, { op: "add", path: this.path, value: _deepClone(valueToCopy) });
      return { newDocument: document2 };
    },
    test: function(obj, key, document2) {
      return { newDocument: document2, test: _areEquals(obj[key], this.value) };
    },
    _get: function(obj, key, document2) {
      this.value = obj[key];
      return { newDocument: document2 };
    }
  };
  var arrOps = {
    add: function(arr, i, document2) {
      if (isInteger(i)) {
        arr.splice(i, 0, this.value);
      } else {
        arr[i] = this.value;
      }
      return { newDocument: document2, index: i };
    },
    remove: function(arr, i, document2) {
      var removedList = arr.splice(i, 1);
      return { newDocument: document2, removed: removedList[0] };
    },
    replace: function(arr, i, document2) {
      var removed = arr[i];
      arr[i] = this.value;
      return { newDocument: document2, removed };
    },
    move: objOps.move,
    copy: objOps.copy,
    test: objOps.test,
    _get: objOps._get
  };
  function getValueByPointer(document2, pointer) {
    if (pointer == "") {
      return document2;
    }
    var getOriginalDestination = { op: "_get", path: pointer };
    applyOperation(document2, getOriginalDestination);
    return getOriginalDestination.value;
  }
  function applyOperation(document2, operation, validateOperation, mutateDocument, banPrototypeModifications, index) {
    if (validateOperation === void 0) {
      validateOperation = false;
    }
    if (mutateDocument === void 0) {
      mutateDocument = true;
    }
    if (banPrototypeModifications === void 0) {
      banPrototypeModifications = true;
    }
    if (index === void 0) {
      index = 0;
    }
    if (validateOperation) {
      if (typeof validateOperation == "function") {
        validateOperation(operation, 0, document2, operation.path);
      } else {
        validator(operation, 0);
      }
    }
    if (operation.path === "") {
      var returnValue = { newDocument: document2 };
      if (operation.op === "add") {
        returnValue.newDocument = operation.value;
        return returnValue;
      } else if (operation.op === "replace") {
        returnValue.newDocument = operation.value;
        returnValue.removed = document2;
        return returnValue;
      } else if (operation.op === "move" || operation.op === "copy") {
        returnValue.newDocument = getValueByPointer(document2, operation.from);
        if (operation.op === "move") {
          returnValue.removed = document2;
        }
        return returnValue;
      } else if (operation.op === "test") {
        returnValue.test = _areEquals(document2, operation.value);
        if (returnValue.test === false) {
          throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
        }
        returnValue.newDocument = document2;
        return returnValue;
      } else if (operation.op === "remove") {
        returnValue.removed = document2;
        returnValue.newDocument = null;
        return returnValue;
      } else if (operation.op === "_get") {
        operation.value = document2;
        return returnValue;
      } else {
        if (validateOperation) {
          throw new JsonPatchError("Operation `op` property is not one of operations defined in RFC-6902", "OPERATION_OP_INVALID", index, operation, document2);
        } else {
          return returnValue;
        }
      }
    } else {
      if (!mutateDocument) {
        document2 = _deepClone(document2);
      }
      var path = operation.path || "";
      var keys = path.split("/");
      var obj = document2;
      var t2 = 1;
      var len = keys.length;
      var existingPathFragment = void 0;
      var key = void 0;
      var validateFunction = void 0;
      if (typeof validateOperation == "function") {
        validateFunction = validateOperation;
      } else {
        validateFunction = validator;
      }
      while (true) {
        key = keys[t2];
        if (key && key.indexOf("~") != -1) {
          key = unescapePathComponent(key);
        }
        if (banPrototypeModifications && (key == "__proto__" || key == "prototype" && t2 > 0 && keys[t2 - 1] == "constructor")) {
          throw new TypeError("JSON-Patch: modifying `__proto__` or `constructor/prototype` prop is banned for security reasons, if this was on purpose, please set `banPrototypeModifications` flag false and pass it to this function. More info in fast-json-patch README");
        }
        if (validateOperation) {
          if (existingPathFragment === void 0) {
            if (obj[key] === void 0) {
              existingPathFragment = keys.slice(0, t2).join("/");
            } else if (t2 == len - 1) {
              existingPathFragment = operation.path;
            }
            if (existingPathFragment !== void 0) {
              validateFunction(operation, 0, document2, existingPathFragment);
            }
          }
        }
        t2++;
        if (Array.isArray(obj)) {
          if (key === "-") {
            key = obj.length;
          } else {
            if (validateOperation && !isInteger(key)) {
              throw new JsonPatchError("Expected an unsigned base-10 integer value, making the new referenced value the array element with the zero-based index", "OPERATION_PATH_ILLEGAL_ARRAY_INDEX", index, operation, document2);
            } else if (isInteger(key)) {
              key = ~~key;
            }
          }
          if (t2 >= len) {
            if (validateOperation && operation.op === "add" && key > obj.length) {
              throw new JsonPatchError("The specified index MUST NOT be greater than the number of elements in the array", "OPERATION_VALUE_OUT_OF_BOUNDS", index, operation, document2);
            }
            var returnValue = arrOps[operation.op].call(operation, obj, key, document2);
            if (returnValue.test === false) {
              throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
            }
            return returnValue;
          }
        } else {
          if (t2 >= len) {
            var returnValue = objOps[operation.op].call(operation, obj, key, document2);
            if (returnValue.test === false) {
              throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
            }
            return returnValue;
          }
        }
        obj = obj[key];
        if (validateOperation && t2 < len && (!obj || typeof obj !== "object")) {
          throw new JsonPatchError("Cannot perform operation at the desired path", "OPERATION_PATH_UNRESOLVABLE", index, operation, document2);
        }
      }
    }
  }
  function applyPatch(document2, patch, validateOperation, mutateDocument, banPrototypeModifications) {
    if (mutateDocument === void 0) {
      mutateDocument = true;
    }
    if (banPrototypeModifications === void 0) {
      banPrototypeModifications = true;
    }
    if (validateOperation) {
      if (!Array.isArray(patch)) {
        throw new JsonPatchError("Patch sequence must be an array", "SEQUENCE_NOT_AN_ARRAY");
      }
    }
    if (!mutateDocument) {
      document2 = _deepClone(document2);
    }
    var results = new Array(patch.length);
    for (var i = 0, length_1 = patch.length; i < length_1; i++) {
      results[i] = applyOperation(document2, patch[i], validateOperation, true, banPrototypeModifications, i);
      document2 = results[i].newDocument;
    }
    results.newDocument = document2;
    return results;
  }
  function applyReducer(document2, operation, index) {
    var operationResult = applyOperation(document2, operation);
    if (operationResult.test === false) {
      throw new JsonPatchError("Test operation failed", "TEST_OPERATION_FAILED", index, operation, document2);
    }
    return operationResult.newDocument;
  }
  function validator(operation, index, document2, existingPathFragment) {
    if (typeof operation !== "object" || operation === null || Array.isArray(operation)) {
      throw new JsonPatchError("Operation is not an object", "OPERATION_NOT_AN_OBJECT", index, operation, document2);
    } else if (!objOps[operation.op]) {
      throw new JsonPatchError("Operation `op` property is not one of operations defined in RFC-6902", "OPERATION_OP_INVALID", index, operation, document2);
    } else if (typeof operation.path !== "string") {
      throw new JsonPatchError("Operation `path` property is not a string", "OPERATION_PATH_INVALID", index, operation, document2);
    } else if (operation.path.indexOf("/") !== 0 && operation.path.length > 0) {
      throw new JsonPatchError('Operation `path` property must start with "/"', "OPERATION_PATH_INVALID", index, operation, document2);
    } else if ((operation.op === "move" || operation.op === "copy") && typeof operation.from !== "string") {
      throw new JsonPatchError("Operation `from` property is not present (applicable in `move` and `copy` operations)", "OPERATION_FROM_REQUIRED", index, operation, document2);
    } else if ((operation.op === "add" || operation.op === "replace" || operation.op === "test") && operation.value === void 0) {
      throw new JsonPatchError("Operation `value` property is not present (applicable in `add`, `replace` and `test` operations)", "OPERATION_VALUE_REQUIRED", index, operation, document2);
    } else if ((operation.op === "add" || operation.op === "replace" || operation.op === "test") && hasUndefined(operation.value)) {
      throw new JsonPatchError("Operation `value` property is not present (applicable in `add`, `replace` and `test` operations)", "OPERATION_VALUE_CANNOT_CONTAIN_UNDEFINED", index, operation, document2);
    } else if (document2) {
      if (operation.op == "add") {
        var pathLen = operation.path.split("/").length;
        var existingPathLen = existingPathFragment.split("/").length;
        if (pathLen !== existingPathLen + 1 && pathLen !== existingPathLen) {
          throw new JsonPatchError("Cannot perform an `add` operation at the desired path", "OPERATION_PATH_CANNOT_ADD", index, operation, document2);
        }
      } else if (operation.op === "replace" || operation.op === "remove" || operation.op === "_get") {
        if (operation.path !== existingPathFragment) {
          throw new JsonPatchError("Cannot perform the operation at a path that does not exist", "OPERATION_PATH_UNRESOLVABLE", index, operation, document2);
        }
      } else if (operation.op === "move" || operation.op === "copy") {
        var existingValue = { op: "_get", path: operation.from, value: void 0 };
        var error = validate([existingValue], document2);
        if (error && error.name === "OPERATION_PATH_UNRESOLVABLE") {
          throw new JsonPatchError("Cannot perform the operation from a path that does not exist", "OPERATION_FROM_UNRESOLVABLE", index, operation, document2);
        }
      }
    }
  }
  function validate(sequence, document2, externalValidator) {
    try {
      if (!Array.isArray(sequence)) {
        throw new JsonPatchError("Patch sequence must be an array", "SEQUENCE_NOT_AN_ARRAY");
      }
      if (document2) {
        applyPatch(_deepClone(document2), _deepClone(sequence), externalValidator || true);
      } else {
        externalValidator = externalValidator || validator;
        for (var i = 0; i < sequence.length; i++) {
          externalValidator(sequence[i], i, document2, void 0);
        }
      }
    } catch (e) {
      if (e instanceof JsonPatchError) {
        return e;
      } else {
        throw e;
      }
    }
  }
  function _areEquals(a, b) {
    if (a === b)
      return true;
    if (a && b && typeof a == "object" && typeof b == "object") {
      var arrA = Array.isArray(a), arrB = Array.isArray(b), i, length, key;
      if (arrA && arrB) {
        length = a.length;
        if (length != b.length)
          return false;
        for (i = length; i-- !== 0; )
          if (!_areEquals(a[i], b[i]))
            return false;
        return true;
      }
      if (arrA != arrB)
        return false;
      var keys = Object.keys(a);
      length = keys.length;
      if (length !== Object.keys(b).length)
        return false;
      for (i = length; i-- !== 0; )
        if (!b.hasOwnProperty(keys[i]))
          return false;
      for (i = length; i-- !== 0; ) {
        key = keys[i];
        if (!_areEquals(a[key], b[key]))
          return false;
      }
      return true;
    }
    return a !== a && b !== b;
  }

  // node_modules/fast-json-patch/module/duplex.mjs
  var duplex_exports = {};
  __export(duplex_exports, {
    compare: () => compare,
    generate: () => generate,
    observe: () => observe,
    unobserve: () => unobserve
  });
  var beforeDict = /* @__PURE__ */ new WeakMap();
  var Mirror = (
    /** @class */
    /* @__PURE__ */ (function() {
      function Mirror2(obj) {
        this.observers = /* @__PURE__ */ new Map();
        this.obj = obj;
      }
      return Mirror2;
    })()
  );
  var ObserverInfo = (
    /** @class */
    /* @__PURE__ */ (function() {
      function ObserverInfo2(callback, observer) {
        this.callback = callback;
        this.observer = observer;
      }
      return ObserverInfo2;
    })()
  );
  function getMirror(obj) {
    return beforeDict.get(obj);
  }
  function getObserverFromMirror(mirror, callback) {
    return mirror.observers.get(callback);
  }
  function removeObserverFromMirror(mirror, observer) {
    mirror.observers.delete(observer.callback);
  }
  function unobserve(root, observer) {
    observer.unobserve();
  }
  function observe(obj, callback) {
    var patches = [];
    var observer;
    var mirror = getMirror(obj);
    if (!mirror) {
      mirror = new Mirror(obj);
      beforeDict.set(obj, mirror);
    } else {
      var observerInfo = getObserverFromMirror(mirror, callback);
      observer = observerInfo && observerInfo.observer;
    }
    if (observer) {
      return observer;
    }
    observer = {};
    mirror.value = _deepClone(obj);
    if (callback) {
      observer.callback = callback;
      observer.next = null;
      var dirtyCheck = function() {
        generate(observer);
      };
      var fastCheck = function() {
        clearTimeout(observer.next);
        observer.next = setTimeout(dirtyCheck);
      };
      if (typeof window !== "undefined") {
        window.addEventListener("mouseup", fastCheck);
        window.addEventListener("keyup", fastCheck);
        window.addEventListener("mousedown", fastCheck);
        window.addEventListener("keydown", fastCheck);
        window.addEventListener("change", fastCheck);
      }
    }
    observer.patches = patches;
    observer.object = obj;
    observer.unobserve = function() {
      generate(observer);
      clearTimeout(observer.next);
      removeObserverFromMirror(mirror, observer);
      if (typeof window !== "undefined") {
        window.removeEventListener("mouseup", fastCheck);
        window.removeEventListener("keyup", fastCheck);
        window.removeEventListener("mousedown", fastCheck);
        window.removeEventListener("keydown", fastCheck);
        window.removeEventListener("change", fastCheck);
      }
    };
    mirror.observers.set(callback, new ObserverInfo(callback, observer));
    return observer;
  }
  function generate(observer, invertible) {
    if (invertible === void 0) {
      invertible = false;
    }
    var mirror = beforeDict.get(observer.object);
    _generate(mirror.value, observer.object, observer.patches, "", invertible);
    if (observer.patches.length) {
      applyPatch(mirror.value, observer.patches);
    }
    var temp = observer.patches;
    if (temp.length > 0) {
      observer.patches = [];
      if (observer.callback) {
        observer.callback(temp);
      }
    }
    return temp;
  }
  function _generate(mirror, obj, patches, path, invertible) {
    if (obj === mirror) {
      return;
    }
    if (typeof obj.toJSON === "function") {
      obj = obj.toJSON();
    }
    var newKeys = _objectKeys(obj);
    var oldKeys = _objectKeys(mirror);
    var changed = false;
    var deleted = false;
    for (var t2 = oldKeys.length - 1; t2 >= 0; t2--) {
      var key = oldKeys[t2];
      var oldVal = mirror[key];
      if (hasOwnProperty(obj, key) && !(obj[key] === void 0 && oldVal !== void 0 && Array.isArray(obj) === false)) {
        var newVal = obj[key];
        if (typeof oldVal == "object" && oldVal != null && typeof newVal == "object" && newVal != null && Array.isArray(oldVal) === Array.isArray(newVal)) {
          _generate(oldVal, newVal, patches, path + "/" + escapePathComponent(key), invertible);
        } else {
          if (oldVal !== newVal) {
            changed = true;
            if (invertible) {
              patches.push({ op: "test", path: path + "/" + escapePathComponent(key), value: _deepClone(oldVal) });
            }
            patches.push({ op: "replace", path: path + "/" + escapePathComponent(key), value: _deepClone(newVal) });
          }
        }
      } else if (Array.isArray(mirror) === Array.isArray(obj)) {
        if (invertible) {
          patches.push({ op: "test", path: path + "/" + escapePathComponent(key), value: _deepClone(oldVal) });
        }
        patches.push({ op: "remove", path: path + "/" + escapePathComponent(key) });
        deleted = true;
      } else {
        if (invertible) {
          patches.push({ op: "test", path, value: mirror });
        }
        patches.push({ op: "replace", path, value: obj });
        changed = true;
      }
    }
    if (!deleted && newKeys.length == oldKeys.length) {
      return;
    }
    for (var t2 = 0; t2 < newKeys.length; t2++) {
      var key = newKeys[t2];
      if (!hasOwnProperty(mirror, key) && obj[key] !== void 0) {
        patches.push({ op: "add", path: path + "/" + escapePathComponent(key), value: _deepClone(obj[key]) });
      }
    }
  }
  function compare(tree1, tree2, invertible) {
    if (invertible === void 0) {
      invertible = false;
    }
    var patches = [];
    _generate(tree1, tree2, patches, "", invertible);
    return patches;
  }

  // node_modules/fast-json-patch/index.mjs
  var fast_json_patch_default = Object.assign({}, core_exports, duplex_exports, {
    JsonPatchError: PatchError,
    deepClone: _deepClone,
    escapePathComponent,
    unescapePathComponent
  });

  // src/core/store.js
  var { applyPatch: applyPatch2 } = fast_json_patch_default;
  function objKey(kind, scope, name) {
    return kind + "|" + (scope || "global") + "|" + (name || "");
  }
  function normalizeRef(x, kindHint) {
    if (x == null) return null;
    if (typeof x === "string") {
      return kindHint ? { kind: String(kindHint), scope: "global", name: x, rv: 0, raw: x } : null;
    }
    if (typeof x !== "object") return null;
    const kind = x.kind || kindHint;
    const name = x.name;
    if (!kind || !name) return null;
    const rvRaw = x.resourceVersion != null ? x.resourceVersion : x.rv;
    return { kind: String(kind), scope: String(x.scope || "global"), name: String(name), rv: Number(rvRaw) || 0, raw: x };
  }
  function arrOf(v) {
    return Array.isArray(v) ? v : [];
  }
  function cloneSnapshot(v) {
    try {
      return typeof structuredClone === "function" ? structuredClone(v) : JSON.parse(JSON.stringify(v));
    } catch {
      return v;
    }
  }
  function createStore() {
    let state = null;
    let seq2 = 0;
    const subs2 = /* @__PURE__ */ new Set();
    let scheduled = false;
    const docCache = /* @__PURE__ */ new Map();
    function notify() {
      if (scheduled) return;
      scheduled = true;
      queueMicrotask(() => {
        scheduled = false;
        subs2.forEach((fn) => {
          try {
            fn(state);
          } catch (e) {
            console.error("[store] \u8BA2\u9605\u8005\u5931\u8D25", e);
          }
        });
      });
    }
    const api2 = {
      /** 快照是否就绪（未 init 前=骨架词形）。 */
      get isReady() {
        return !!state;
      },
      /** 建基线（整段快照——首连/patch 失败回退/降级轮询共用同一入口）。 */
      init(snapshot) {
        state = snapshot && typeof snapshot === "object" ? cloneSnapshot(snapshot) : {};
        if (Number.isFinite(Number(state.seq))) seq2 = Number(state.seq);
        notify();
        return api2;
      },
      /**
       * 增量应用 RFC6902 patch（validate:false 保守应用）。
       * 返回 true=已应用并通知；false=应用失败（状态保持原样——调用方整段 re-fetch）。
       */
      applyPatch(patch, meta) {
        if (!state) return false;
        if (!Array.isArray(patch) || !patch.length) return true;
        let ok = true;
        try {
          const res = applyPatch2(state, patch, false, true);
          if (res && res.newDocument) state = res.newDocument;
        } catch (e) {
          ok = false;
          console.warn("[store] patch \u5E94\u7528\u5931\u8D25\u2014\u2014\u56DE\u9000\u6574\u6BB5\u5FEB\u7167\uFF1A", String(e && e.message || e));
        }
        if (!ok) return false;
        if (meta && Number.isFinite(Number(meta.seq))) seq2 = Number(meta.seq);
        notify();
        return true;
      },
      /** 最近应用的 patch 序号。 */
      get seq() {
        return seq2;
      },
      /** 连接态（connecting/connected/reconnecting/degraded——ws.js 广播，随状态帧通知）。 */
      setConn(s) {
        if (conn !== s) {
          conn = s;
          notify();
        }
      },
      get conn() {
        return conn;
      },
      /** 点路径读（'a.b.0.c'——缺路径=undefined 不抛）。 */
      get(path) {
        if (!state) return void 0;
        return String(path || "").split(".").reduce((o, k) => o == null ? void 0 : o[k], state);
      },
      /** 订阅状态帧（返回退订函数）。 */
      subscribe(fn) {
        subs2.add(fn);
        return () => subs2.delete(fn);
      },
      /* ---------- 容器化访问器（快照字段缺席=空数组/空值，不抛不猜） ---------- */
      objects() {
        const o = state && state.objects;
        if (!o) return [];
        if (Array.isArray(o)) return o.map((x) => normalizeRef(x)).filter(Boolean);
        return Object.keys(o).flatMap((k) => arrOf(o[k]).map((x) => normalizeRef(x, k))).filter(Boolean);
      },
      refs(kind) {
        return api2.objects().filter((x) => x.kind === kind);
      },
      kinds() {
        return arrOf(state && state.kinds);
      },
      events() {
        return arrOf(state && state.events);
      },
      drafts() {
        return arrOf(state && state.drafts);
      },
      tasks() {
        const t2 = state && state.tasks;
        if (Array.isArray(t2)) return t2;
        return arrOf(t2 && t2.tasks);
      },
      runs() {
        const r = state && (state.runs !== void 0 ? state.runs : state.workflowRuns);
        if (Array.isArray(r)) return r;
        if (r && Array.isArray(r.recent)) return r.recent;
        return arrOf(r && r.runs);
      },
      workflows() {
        const w = state && state.workflows;
        return Array.isArray(w) ? w : arrOf(w && w.workflows);
      },
      proposals() {
        const p = state && state.proposals;
        return Array.isArray(p) ? p : arrOf(p && p.proposals);
      },
      sessions() {
        const s = state && state.sessions;
        return Array.isArray(s) ? s : arrOf(s && s.sessions);
      },
      attention() {
        const a = state && state.attention;
        if (!a) return [];
        if (Array.isArray(a)) return a;
        return arrOf(a.items || a.cards);
      },
      inbox() {
        return state && state.inbox || null;
      },
      /* ---------- 对象级 rv 缓存（正文懒取的家） ---------- */
      cacheDoc(kind, scope, name, rv, content) {
        docCache.set(objKey(kind, scope, name), {
          rv: Number(rv) || 0,
          content: content == null ? "" : String(content),
          at: Date.now()
        });
      },
      /** 命中判据：rv 未变才返回缓存（rv 变了=过期不返回——调用方现读）。 */
      cachedDoc(kind, scope, name, rv) {
        const c = docCache.get(objKey(kind, scope, name));
        if (!c) return null;
        const want = Number(rv) || 0;
        if (want > 0 && c.rv !== want) return null;
        return c;
      },
      /** 快照中 rv 与缓存不符（或未缓存）的对象 refs——首连/降级轮询后正文预取清单。 */
      staleDocs(limit) {
        const cap = Number(limit) || 24;
        const out = [];
        for (const r of api2.objects()) {
          if (!r.kind || !r.name) continue;
          const c = docCache.get(objKey(r.kind, r.scope, r.name));
          if (!c || c.rv !== r.rv) {
            out.push(r);
            if (out.length >= cap) break;
          }
        }
        return out;
      },
      clearDocCache() {
        docCache.clear();
      }
    };
    let conn = "connecting";
    return api2;
  }

  // src/core/api.js
  function createApi(base = "", paths = {}) {
    const B = String(base || "").replace(/\/+$/, "");
    const enc = encodeURIComponent;
    const P = {
      apply: paths.apply || "v1/apply",
      drafts: paths.drafts || "v1/drafts"
    };
    async function request(path, opts = {}) {
      const p = String(path);
      const url = p.startsWith("/") ? p : B + "/" + p.replace(/^\/+/, "");
      try {
        const res = await fetch(url, {
          method: opts.method || "GET",
          headers: opts.body !== void 0 ? { "Content-Type": "application/json" } : void 0,
          body: opts.body !== void 0 ? JSON.stringify(opts.body) : void 0,
          credentials: "same-origin"
        });
        const text2 = await res.text();
        let data = null;
        try {
          data = text2 ? JSON.parse(text2) : null;
        } catch {
          data = { raw: text2 };
        }
        if (!res.ok) {
          const err = data && (data.error || data.message);
          return { ok: false, status: res.status, error: err ? String(err) : "HTTP " + res.status };
        }
        return { ok: true, data };
      } catch (e) {
        return { ok: false, status: 0, error: String(e && e.message || e) };
      }
    }
    return {
      request,
      /* ---------- 新契约端点 ---------- */
      snapshot: () => request("v1/snapshot"),
      event: (id) => request("v1/events/" + enc(String(id == null ? "" : id))),
      loopCurrent: () => request("v1/loop/current"),
      /* ---------- 既有只读面 ---------- */
      /* R1（§19.2）：scope 恒显式（cored 词形「写入/读取需显式 scope」——global 亦不省，
       * 省略=400「scope 与 storage 模板不相容」噪声 15 条/页+正文恒注水中）。 */
      object: (kind, scope, name) => request("v1/objects/" + enc(kind) + "/" + enc(name) + "?scope=" + enc(scope || "global")),
      kinds: () => request("v1/scheme/kinds"),
      drafts: () => request("v1/drafts"),
      draft: (kind, name, scope) => request("v1/drafts/" + enc(kind) + "/" + enc(name) + (scope && scope !== "global" ? "?scope=" + enc(scope) : "")),
      events: (limit = 200) => request("v1/events?limit=" + Number(limit)),
      workflows: () => request("v1/workflows"),
      workflowRuns: () => request("v1/workflows/runs"),
      sessions: () => request("v1/sessions"),
      session: (id) => request("v1/sessions/" + enc(String(id == null ? "" : id))),
      /* 会话上下文契约面（G-14——routes_sessions.go：ContextPack 词形投影+编辑分叉） */
      sessionContext: (id) => request("v1/sessions/" + enc(String(id == null ? "" : id)) + "/context"),
      saveSessionContext: (id, payload) => request("v1/sessions/" + enc(String(id == null ? "" : id)) + "/context", { method: "PUT", body: payload }),
      revertSessionContext: (id) => request("v1/sessions/" + enc(String(id == null ? "" : id)) + "/context/revert", { method: "POST" }),
      inbox: () => request("v1/inbox"),
      proposals: () => request("v1/proposals"),
      approvals: (limit = 20) => request("v1/approvals?limit=" + Number(limit)),
      designDocs: () => request("v1/design-docs"),
      designDoc: (repo, path) => request("v1/design-docs/" + enc(repo) + "/" + enc(path)),
      /* ---------- 写通路（唯一入口——客户端零旁路） ---------- */
      apply: (payload) => request(P.apply, { method: "POST", body: payload }),
      saveDraft: (payload) => request(P.drafts, { method: "POST", body: payload }),
      /* 事件注入通路（§19 R3——人工批闸/点火/取消族；部署=宿主路由 /core-console/event-publish
       * （绝对词位——经 plugin 注入主体，source=human 服务端强制）。 */
      eventPub: (payload) => request("/core-console/event-pub", { method: "POST", body: payload })
    };
  }

  // src/core/router.js
  function parseHash(hash) {
    let h = String(hash || "").replace(/^#\/?/, "");
    h = h.split("?")[0];
    if (h === "") return [];
    return h.split("/").map((s) => {
      try {
        return decodeURIComponent(s);
      } catch {
        return s;
      }
    });
  }
  function matchSegs(pattern, segs) {
    const ps = String(pattern).split("/");
    if (ps.length !== segs.length) return null;
    const params = {};
    for (let i = 0; i < ps.length; i++) {
      const p = ps[i];
      const s = segs[i];
      if (p.charAt(0) === ":") params[p.slice(1)] = s;
      else if (p !== s) return null;
    }
    return params;
  }
  function createRouter(opts = {}) {
    const routes = [];
    let viewEl = null;
    let current = { route: "", pattern: "", segs: [], params: {} };
    function resolve(segs) {
      for (const r of routes) {
        const params = matchSegs(r.pattern, segs);
        if (params) return { r, params };
      }
      return null;
    }
    function notFoundHtml(segs) {
      return '<div class="view-inner"><div class="empty"><div class="e-title">\u672A\u77E5\u8DEF\u5F84\uFF08' + segs.map((s) => s).join("/") + '\uFF09</div><div class="small faint">\u8DEF\u7531\u8BCD\u5F62\u89C1\u63A7\u5236\u53F0\u5BFC\u822A\u2014\u2014\u6216\u4ECE Welcome \u8FDB\u5165\u3002</div><div class="e-act"><button class="btn primary" data-go="welcome">\u56DE Welcome</button></div></div></div>';
    }
    const api2 = {
      /** 注册屏（pattern 如 'wf/:name/r/:n'；fn(viewEl, params, ctx)）。 */
      route(pattern, fn) {
        routes.push({ pattern: String(pattern), fn });
        return api2;
      },
      resolve,
      get current() {
        return current;
      },
      /** 导航（location.hash 变化触发 render；同址=强制重渲染）。 */
      go(path) {
        if (typeof location === "undefined") return;
        const h = "#/" + String(path || "").replace(/^\/+/, "");
        if (location.hash === h) api2.render();
        else location.hash = h;
      },
      /** 装配：挂容器 + 监听 hashchange + 首渲染。 */
      mount(el) {
        viewEl = el;
        if (typeof window !== "undefined") {
          window.addEventListener("hashchange", () => api2.render());
        }
        api2.render();
        return api2;
      },
      /** 解析当前 hash → 匹配屏执行（未命中=404 词形，不冒充内容）。 */
      render() {
        if (!viewEl || typeof location === "undefined") return null;
        const segs = parseHash(location.hash);
        if (!segs.length) segs.push(opts.defaultRoute || "welcome");
        const hit = resolve(segs);
        if (!hit) {
          current = { route: "404", pattern: "", segs, params: {} };
          viewEl.innerHTML = notFoundHtml(segs);
          return null;
        }
        current = { route: hit.r.pattern.split("/")[0], pattern: hit.r.pattern, segs, params: hit.params };
        hit.r.fn(viewEl, hit.params);
        return hit.r;
      }
    };
    return api2;
  }

  // src/core/ws.js
  function wsUrlOf(loc, path = "v1/stream") {
    const l = loc || (typeof location !== "undefined" ? location : null);
    if (!l || !l.host) return "/" + String(path).replace(/^\/+/, "");
    const proto = l.protocol === "https:" ? "wss://" : "ws://";
    return proto + l.host + "/" + String(path).replace(/^\/+/, "");
  }
  function hidden() {
    return typeof document !== "undefined" && !!document.hidden;
  }
  function connectStream(opts = {}) {
    const {
      url = wsUrlOf(),
      base = "",
      onSnapshot,
      onPatch,
      onState,
      pollMs = 3e4,
      maxFail = 5
    } = opts;
    const B = String(base || "").replace(/\/+$/, "");
    let ws = null;
    let state = "connecting";
    let fails = 0;
    let backoff = 1e3;
    let timer = null;
    let pollTimer = null;
    let pendingReconnect = false;
    let closed = false;
    function setState(s) {
      if (state === s) return;
      state = s;
      try {
        if (onState) onState(s);
      } catch {
      }
    }
    function clearTimers() {
      if (timer) {
        clearTimeout(timer);
        timer = null;
      }
      if (pollTimer) {
        clearInterval(pollTimer);
        pollTimer = null;
      }
    }
    function pollOnce() {
      if (closed || hidden()) return;
      fetch(B + "/v1/snapshot", { credentials: "same-origin" }).then((r) => r.ok ? r.json() : null).then((j) => {
        if (!j || closed) return;
        fails = 0;
        try {
          if (onSnapshot) onSnapshot(j);
        } catch {
        }
      }).catch(() => {
      });
    }
    function startPolling() {
      setState("degraded");
      if (pollTimer) return;
      pollOnce();
      pollTimer = setInterval(pollOnce, pollMs);
    }
    function scheduleReconnect() {
      if (closed) return;
      if (hidden()) {
        pendingReconnect = true;
        return;
      }
      if (timer) return;
      timer = setTimeout(() => {
        timer = null;
        open();
      }, backoff);
      backoff = Math.min(backoff * 2, 3e4);
    }
    function open() {
      if (closed) return;
      setState(fails > 0 ? "reconnecting" : "connecting");
      let sock;
      try {
        sock = new WebSocket(url);
      } catch {
        onFailure();
        return;
      }
      ws = sock;
      sock.onopen = () => {
        fails = 0;
        backoff = 1e3;
        if (pollTimer) {
          clearInterval(pollTimer);
          pollTimer = null;
        }
        setState("connected");
      };
      sock.onmessage = (ev) => {
        let msg = null;
        try {
          msg = JSON.parse(String(ev.data));
        } catch {
          return;
        }
        if (!msg || !msg.type) return;
        if (msg.type === "snapshot") {
          try {
            if (onSnapshot) onSnapshot(msg.snapshot !== void 0 ? msg.snapshot : msg.data !== void 0 ? msg.data : msg);
          } catch {
          }
        } else if (msg.type === "patch") {
          try {
            if (onPatch) onPatch(Array.isArray(msg.patch) ? msg.patch : [], msg);
          } catch {
          }
        } else if (msg.type === "ping") {
          try {
            if (sock.readyState === 1) sock.send(JSON.stringify({ type: "pong" }));
          } catch {
          }
        }
      };
      sock.onclose = () => {
        if (ws === sock) {
          ws = null;
          onFailure();
        }
      };
      sock.onerror = () => {
      };
    }
    function onFailure() {
      if (closed) return;
      fails += 1;
      if (fails >= maxFail) startPolling();
      else scheduleReconnect();
    }
    function resume() {
      if (closed) return;
      if (pendingReconnect) {
        pendingReconnect = false;
        scheduleReconnect();
        return;
      }
      if (ws) {
        if (state === "degraded") pollOnce();
        return;
      }
      if (state === "degraded") {
        backoff = 1e3;
        open();
        return;
      }
      if (state !== "connected") {
        open();
      }
    }
    function onVisibility() {
      if (hidden()) {
        if (timer) {
          clearTimeout(timer);
          timer = null;
          pendingReconnect = true;
        }
      } else {
        resume();
      }
    }
    if (typeof document !== "undefined") {
      document.addEventListener("visibilitychange", onVisibility);
    }
    open();
    return {
      close() {
        closed = true;
        clearTimers();
        if (typeof document !== "undefined") document.removeEventListener("visibilitychange", onVisibility);
        if (ws) {
          try {
            ws.close();
          } catch {
          }
          ws = null;
        }
      },
      get state() {
        return state;
      },
      send(obj) {
        if (ws && ws.readyState === 1) {
          try {
            ws.send(JSON.stringify(obj));
            return true;
          } catch {
            return false;
          }
        }
        return false;
      }
    };
  }

  // src/render/chips.js
  function chip(label, cls = "mute", tiny = true, attrs = "") {
    return '<span class="chip ' + esc(cls) + (tiny ? " tiny" : "") + '"' + (attrs ? " " + attrs : "") + ">" + esc(label) + "</span>";
  }
  function chipGo(label, route, cls = "mute", tiny = true, title = "") {
    return '<span class="chip ' + esc(cls) + (tiny ? " tiny" : "") + ' click" data-go="' + esc(route) + '"' + (title ? ' title="' + esc(title) + '"' : "") + ">" + esc(label) + "</span>";
  }
  function kindBadge(kind) {
    return '<span class="chip kind tiny">' + esc(kind) + "</span>";
  }
  function kv(rows) {
    return '<div class="kv">' + (rows || []).filter((r) => r && r.length === 2).map(
      ([k, v]) => '<span class="k">' + esc(k) + '</span><span class="v">' + v + "</span>"
    ).join("") + "</div>";
  }
  function emptyBox(title, sub = "", actionHtml = "") {
    return '<div class="empty"><div class="e-title">' + esc(title) + "</div>" + (sub ? '<div class="small faint">' + esc(sub) + "</div>" : "") + (actionHtml ? '<div class="e-act">' + actionHtml + "</div>" : "") + "</div>";
  }
  function screenHead(title, sub = "", actionsHtml = "") {
    return '<div class="dh"><div class="dh-main"><div class="dh-title"><h1>' + esc(title) + "</h1></div>" + (sub ? '<div class="dh-sub">' + esc(sub) + "</div>" : "") + "</div>" + (actionsHtml ? '<div class="dh-actions">' + actionsHtml + "</div>" : "") + "</div>";
  }
  function encSeg(s) {
    return encodeURIComponent(String(s == null ? "" : s));
  }
  function objRoute(kind, scope, name) {
    return "obj/" + encSeg(kind) + "/" + encSeg(scope || "global") + "/" + encSeg(name);
  }
  function docRoute(kind, scope, name) {
    return "doc/" + encSeg(kind) + "/" + encSeg(scope || "global") + "/" + encSeg(name);
  }
  function involvedRoute2(s) {
    const r = String(s || "").trim();
    if (!r) return "";
    if (r.startsWith("#/")) {
      const segs = r.slice(2).split("/").filter(Boolean);
      return segs.length ? segs.map(encSeg).join("/") : "";
    }
    const m = /^([A-Za-z][A-Za-z0-9]*)\/([^/\s]+)\/([^/\s]+)$/.exec(r);
    if (m) return "obj/" + m.slice(1).map(encSeg).join("/");
    return "";
  }
  function syncBadge(runs, name) {
    const active = (runs || []).some((r) => {
      const st = String(r && (r.status || r.state) || "");
      if (!st || st === "done" || st === "failed" || st === "canceled") return false;
      let hay = "";
      try {
        hay = JSON.stringify([r.involved, r.target, r.route, r.workflow]);
      } catch {
        hay = "";
      }
      return hay.includes(name);
    });
    return active ? chip("Running", "acc") : chip("InSync", "ok");
  }

  // src/core/i18n.js
  var KEY = "cc.locale";
  var locale = "zh";
  try {
    locale = typeof localStorage !== "undefined" && localStorage.getItem(KEY) || "zh";
  } catch {
  }
  var en = /* @__PURE__ */ new Map([
    /* 侧栏/壳（main.js） */
    ["Core Console", "Core Console"],
    ["web \u91CD\u6784\u8D77\u70B9", "web rebuild baseline"],
    ["\u8FDE\u63A5\u4E2D\u2026", "Connecting\u2026"],
    ["\u5B9E\u65F6\uFF08WS patch\uFF09", "Live (WS patch)"],
    ["\u91CD\u8FDE\u4E2D\u2026", "Reconnecting\u2026"],
    ["\u964D\u7EA7\u8F6E\u8BE2\uFF0830s\uFF09", "Degraded polling (30s)"],
    ["\u4E2D\u6587", "\u4E2D\u6587"],
    ["EN", "EN"],
    /* welcome */
    ["\u5FEB\u7167\u88C5\u8F7D\u4E2D\u2026", "Loading snapshot\u2026"],
    ["\u6B63\u5728\u5EFA\u7ACB /v1/snapshot \u57FA\u7EBF\uFF08\u8FDE\u63A5\u5EFA\u7ACB\u5373\u63A8\u5168\u91CF\uFF09", "Establishing /v1/snapshot baseline (full push on connect)"],
    ["\u5728\u518C\u5BF9\u8C61", "Registered objects"],
    ["\u5168\u90E8 kind \u5408\u8BA1", "All kinds total"],
    ["kinds", "kinds"],
    ["scheme \u6CE8\u518C", "scheme registry"],
    ["\u4E8B\u4EF6\uFF08\u7A97\u53E3\u5185\uFF09", "Events (window)"],
    ["\u5F52\u6863\u884C", "archived rows"],
    ["\u5F85\u5BA1", "Pending review"],
    ["\u5F85\u5BA1\u533A\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09", "No pending reviews (explicit empty state)"],
    ["\u6536\u4EF6\u7BB1", "Inbox"],
    ["\u5F85\u4F60\u5904\u7406", "awaiting you"],
    ["\u8FDB\u884C\u4E2D\u4EFB\u52A1", "Running tasks"],
    ["\u4EFB\u52A1\u4E2D\u5FC3", "Task Center"],
    ["kinds \u6458\u8981", "kinds summary"],
    ["\u9996\u4E2A\uFF1A", "first: "],
    [" \u4E2A", " item(s)"],
    /* events */
    ["\u4E8B\u4EF6\u6D41", "Event stream"],
    ["\u5F52\u6863\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408 + patch \u589E\u91CF\uFF1B\u70B9\u51FB\u770B\u5168\u6587\u4E0E\u94FE\u8DEF\uFF09", "Archived events (snapshot aggregate + patch increments; click for detail & chain)"],
    ["\u5168\u90E8\u7EA7\u522B", "All levels"],
    ["\u5FAA\u73AF\u5524\u9192", "Loop wake"],
    ["\u4E8B\u4EF6\u8BE6\u60C5", "Event detail"],
    ["\u539F\u59CB\u8F7D\u8377\uFF08JSON\uFF09", "Raw payload (JSON)"],
    ["\u2190 \u56DE\u4E8B\u4EF6\u6D41", "\u2190 Back to events"],
    ["\u5173\u8054", "Linked"],
    ["\u65E0", "none"],
    /* tasks */
    ["\u5206\u7EC4=\u786E\u5B9A\u6027\u6295\u5F71\uFF08\u72B6\u6001\u5355\u4E00\u5199\u8005=\u8C03\u5EA6\u5668\uFF09\uFF1B\u70B9\u51FB\u884C\u8FDB\u8BE6\u60C5", "Groups = deterministic projection (single status writer = scheduler); click a row for detail"],
    ["\u6B63\u5728\u505A", "Running"],
    ["\u7B49\u4EBA", "Waiting on human"],
    ["\u5361\u4F4F", "Stuck"],
    ["\u5B8C\u6210", "Done"],
    ["\u6392\u961F", "Queued"],
    ["\u9879", "item(s)"],
    ["\u6B63\u5728\u505A\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Running: none (explicit empty state)."],
    ["\u7B49\u4EBA\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Waiting on human: none (explicit empty state)."],
    ["\u5361\u4F4F\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Stuck: none (explicit empty state)."],
    ["\u5B8C\u6210\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Done: none (explicit empty state)."],
    ["\u6392\u961F\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Queued: none (explicit empty state)."],
    ["\u72B6\u6001\u672A\u77E5", "state unknown"],
    ["\u6700\u8FD1\u52A8\u4F5C\uFF1A", "last action: "],
    ["\u5173\u6CE8\u62A5\u544A", "Attention Report"],
    ["\u4EFB\u52A1\u8BE6\u60C5", "Task detail"],
    ["\u2190 \u56DE\u4EFB\u52A1\u4E2D\u5FC3", "\u2190 Back to tasks"],
    ["\u8BC1\u636E\u94FE", "Evidence chain"],
    /* objects */
    ["\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "Object Workbench"],
    ["\u8BE5 kind \u6682\u65E0\u5BF9\u8C61", "No objects for this kind"],
    ["\u6682\u65E0\u5728\u518C\u5BF9\u8C61", "No registered objects"],
    ["\u5FAA\u73AF", "Loop"],
    ["\u76F8\u4F4D", "Phase"],
    ["\u8D1F\u8D23", "Owner"],
    ["\u5F00\u59CB\u4E8E", "started at"],
    ["\u76F8\u4F4D\u673A", "Phase machine"],
    ["\u4E8B\u4EF6\u94FE", "Event chain"],
    ["\u5BF9\u8C61\u8BE6\u60C5", "Object detail"],
    ["\u6807\u7B7E", "Tags"],
    ["\u8865\u6807", "Add tag"],
    ["\u4E0A\u4E0B\u6587", "Context"],
    ["\u5DE5\u4F5C\u53F0", "Workbench"],
    ["\u5386\u53F2", "History"],
    ["\u6B63\u6587", "Content"],
    ["\u7F16\u8F91", "Edit"],
    ["\u4FDD\u5B58", "Save"],
    ["\u53D6\u6D88", "Cancel"],
    ["\u8349\u7A3F", "Draft"],
    ["\u5E94\u7528", "Apply"],
    ["\u5F03\u7A3F", "Discard draft"],
    ["\u4E50\u89C2\u9501", "optimistic lock"],
    /* chat / sessionctx */
    ["\u4F1A\u8BDD", "Sessions"],
    ["\u4F1A\u8BDD\u6E05\u5355", "Session list"],
    ["\u4E0E root \u5BF9\u8BDD", "Chat with root"],
    ["\u5BF9\u8BDD", "Chat"],
    ["\u8F6C\u5F55", "Transcript"],
    ["\u8F6C\u5F55\u9762\u672A\u88C5\u914D\u2014\u2014\u5BBF\u4E3B\u6CE8\u5165 window.__CC.chatFaces \u540E\u73B0\uFF1B\u4F1A\u8BDD\u539F\u6587\u4E0D\u5165\u516C\u5171\u5E93\uFF08intake \u6682\u5047\u9762\u4E3A\u51C6\uFF09\u3002", "Transcript face not mounted \u2014 appears after host injects window.__CC.chatFaces; raw session text stays out of the public library (intake placeholder for now)."],
    ["\u5165\u53E3", "Entry points"],
    ["\u76F8\u5173\u4F1A\u8BDD", "Related sessions"],
    ["\u4E0A\u4E0B\u6587\uFF08ContextPack \u5951\u7EA6\u9762\uFF09", "Context (ContextPack contract face)"],
    ["\u7D22\u5F15\u5019\u9009\u6295\u5F71", "index-candidate projection"],
    ["\u5199\u9762\u5F00", "Writes open"],
    ["\u5199\u9762\u5173\u65AD", "Writes closed"],
    ["\u5FC3\u8DF3", "Heartbeats"],
    ["\u6B21 \xB7 \u6700\u8FD1", "hit(s) \xB7 last"],
    ["\u52A8\u4F5C", "Action"],
    ["\u6CE8\u5165\u56FA\u5B9A prompt\u300C\u7EE7\u7EED\u300D\uFF08\u81EA\u6108\u7B2C\u4E00\u52A8\u4F5C\uFF09", "Inject fixed prompt \u300C\u7EE7\u7EED\u300D (self-heal first action)"],
    ["\u7F3A\u7701", "default"],
    ["\u57DF", "range"],
    ["\u672A\u88C5\u914D\uFF08tagSigSource=none\uFF09", "not assembled (tagSigSource=none)"],
    ["\u5206\u7247\u7F13\u5B58", "Shard cache"],
    ["\u5171", "total "],
    ["\u7247 \xB7 stale", " shard(s) \xB7 stale"],
    ["\u91CD\u5EFA\u9608\u503C", "rebuild threshold"],
    ["\u65E0\u8986\u76D6\uFF08\u751F\u6548\u503C=ContextPolicy \u7F3A\u7701\u2014\u2014PUT \u540E\u6B64\u5904\u73B0\u8BFB\u8986\u76D6\u884C\uFF09\u3002", "No override (effective value = ContextPolicy default \u2014 PUT and the override row appears here)."],
    ["\u4FDD\u5B58\u5206\u53C9", "Save fork"],
    ["\u64A4\u9500\u4E0A\u6B21", "Undo last"],
    ["\u5206\u53C9\u5907\u6CE8\uFF08\u53EF\u9009\uFF09", "Fork note (optional)"],
    ["tag \u96C6\uFF1AtagType:tag \u9017\u53F7/\u7A7A\u683C\u5206\u9694\uFF08\u6BCF\u7C7B\u22645 \u603B\u226420\uFF09", "tag set: tagType:tag, comma/space separated (\u22645 per type, \u226420 total)"],
    ["chunks\uFF08\u7D22\u5F15 TopK \u5019\u9009\uFF09", "chunks (index TopK candidates)"],
    ["\u65E0\u7D22\u5F15\u5019\u9009\uFF08tagSig \u672A\u88C5\u914D\u6216\u7D22\u5F15\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u5192\u5145\u88C5\u914D\u5305\u5185\u5BB9\uFF09\u3002", "No index candidates (tagSig not assembled or zero hits \u2014 never fabricates pack content)."],
    ["\u76F8\u5173 Tag \u5206\u7247\uFF08TagCtxShard\uFF09", "Related tag shards (TagCtxShard)"],
    ["\u65E0\u751F\u6548 tag \u5206\u7247\uFF08tagSig \u672A\u88C5\u914D\u2014\u2014\u7F16\u8F91\u5206\u53C9\u540E\u9010 tagKey \u6295\u5F71 TagCtxShard\uFF09\u3002", "No active tag shards (tagSig not assembled \u2014 per-tagKey TagCtxShard projection after fork edit)."],
    ["\u8986\u76D6\u884C", "Override"],
    ["\u672C\u4EE3\u65E0\u4E09\u6BB5\u5F0F Session \u88C5\u914D\u5668\u2014\u2014chunks=\u7D22\u5F15 TopK \u5019\u9009\u6295\u5F71\uFF08context.md \u7D22\u5F15\u7B2C\u4E00\u6BB5\u8BCD\u5F62\uFF09\uFF0C\u975E\u4F1A\u8BDD\u771F\u5B9E\u88C5\u914D\u5305\u3002", "This generation has no three-stage Session assembler \u2014 chunks are index TopK candidate projections (context.md stage-one word form), not the real session pack."],
    /* approvals */
    ["\u5BA1\u6279", "Approvals"],
    ["\u4E24\u6B65\u6279\u51C6", "Two-step approval"],
    ["\u5F85\u5BA1\u533A", "Pending"],
    ["\u5DF2\u51B3\u8D26", "Decision ledger"],
    ["\u6279\u51C6", "Approve"],
    ["\u9A73\u56DE", "Reject"],
    ["\u6682\u65E0\u5F85\u5BA1\u9879\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "No pending items (explicit empty state)."],
    ["\u6682\u65E0\u5DF2\u51B3\u8BB0\u5F55\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "No decisions recorded (explicit empty state)."],
    /* workflows */
    ["\u5DE5\u4F5C\u6D41", "Workflows"],
    ["\u8FD0\u884C", "Run"],
    ["\u8FD0\u884C\u62BD\u5C49", "Run drawer"],
    ["\u89E6\u53D1", "Trigger"],
    ["\u6B65\u9AA4", "Steps"],
    ["\u6682\u65E0\u5DE5\u4F5C\u6D41\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "No workflows (explicit empty state)."],
    /* misc */
    ["\u7CFB\u7EDF\u6811", "System Tree"],
    ["\u6587\u6863\u5E93", "Docs"],
    ["\u63D0\u6848", "Proposals"],
    ["\u8FDE\u63A5\u5C42", "Connectors"],
    ["\u7CFB\u7EDF", "System"],
    ["\u8FDE\u63A5\u5668", "Connectors"],
    ["\u8BBE\u8BA1\u6587\u6863", "Design docs"],
    ["\u6682\u65E0\u5185\u5BB9\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "No content (explicit empty state)."],
    ["\u52A0\u8F7D\u5931\u8D25", "Load failed"],
    ["\u91CD\u8BD5", "Retry"],
    [" vs \u7A3F\u57FA r", " vs draft base r"],
    [" \xB7 \u57DF ", " \xB7 range "],
    [" \xB7 \u5907\u6CE8 ", " \xB7 note "],
    [" \xB7 \u622A\u81F3 ", " \xB7 as of "],
    [" \xB7 \u62A5\u544A\u4E3A\u6D3E\u751F\u89C6\u56FE\u53EF\u968F\u65F6\u91CD\u5EFA\uFF0C\u4E0D\u843D\u5BF9\u8C61\u8D26\u3002</div>", " \xB7 This report is a derived view and can be rebuilt anytime; not persisted as objects.</div>"],
    [" \xB7 \u6700\u8FD1 ", " \xB7 last "],
    [" \xB7 \u68C0\u67E5\u8303\u56F4 ", " \xB7 check scope "],
    [" \xB7 \u7A3F ", " \xB7 draft "],
    [" \xB7 \u7F3A\u5E2D\u6570\u636E\uFF1A", " \xB7 missing data: "],
    [" \xB7 \u91CD\u5EFA\u9608\u503C ", " \xB7 rebuild threshold "],
    [" \u2192 \u5DE5\u4F5C\u6D41", " \u2192 Workflow"],
    [" \u4E2A kind\u2014\u2014\u89C1\u5BF9\u8C61\u5DE5\u4F5C\u53F0\u5DE6\u4FA7\u6E05\u5355\u3002</div>", " kinds \u2014 see the list on the left of the Object Workbench.</div>"],
    [" \u4E2A\u76F8\u5173", " related"],
    [" \u4E2A\uFF1A", " item(s): "],
    [" \u4EFD\u89C4\u5219\u6587\u6863\uFF09\u3002</div>", " rule document(s))."],
    [" \u5904\u53D8\u66F4\uFF09", " change(s))"],
    [" \u603B\u2264", " total \u2264"],
    [" \u6B21", " time(s)"],
    [" \u6B21\u8FD0\u884C", " run(s)"],
    [" \u7247 \xB7 stale ", " shard(s) \xB7 stale "],
    [" \u7684\u5BF9\u8C61\u4E0E\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF09", " objects & events (snapshot aggregate projection)"],
    [" \u7684\u6570\u503C\uFF08ContextPolicy \u9884\u7B97\u57DF\uFF09", " value (ContextPolicy budget range)"],
    [" \u7F3A", " missing"],
    [" \u884C\u2014\u2014\u7528\u8FC7\u6EE4\u6536\u7A84\u3002</div>", " rows \u2014 narrow with the filter.</div>"],
    [" \u884C\u5DEE\u5F02</div>", " line diff</div>"],
    [" \u9879", " item(s)"],
    [" \u9879</span>", " item(s)</span>"],
    ['" data-docs-kind="">\u5168\u90E8</span>', '" data-docs-kind="">All</span>'],
    ['" data-ed-tab="draft">\u8349\u7A3F</div>', '" data-ed-tab="draft">Draft</div>'],
    ['" data-ed-tab="history">\u5386\u53F2</div>', '" data-ed-tab="history">History</div>'],
    ['" data-ed-view-btn="raw">\u539F\u6587</span>', '" data-ed-view-btn="raw">Source</span>'],
    ['" data-ed-view-btn="render">\u6B63\u6587</span>', '" data-ed-view-btn="render">Rendered</span>'],
    ['" title="\u63D0\u53D6\u9884\u7B97\uFF08token \u6570\uFF09">', '" title="Take budget (token count)">'],
    ['" title="\u770B\u540C tag \u5BF9\u8C61">#', '" title="Objects with the same tag">#'],
    ['"><div class="small faint">diff \u73B0\u8BFB\u4E2D\u2026</div></div>', '"><div class="small faint">Loading diff\u2026</div></div>'],
    ['">\u5FFD\u7565</button> ', '">Dismiss</button> '],
    ['">\u5FFD\u7565</button></div></div></div>', '">Dismiss</button></div></div></div>'],
    ['">\u6240\u5C5E scope</button> ', '">Owning scope</button> '],
    ['">\u6279\u51C6\uFF08= Apply\uFF09</button> ', '">Approve (= Apply)</button> '],
    ['">\u7F16\u8F91\u5668</button>', '">Open editor</button>'],
    ['">\u7F16\u8F91\u5668</button> ', '">Open editor</button> '],
    ['</b> <span class="small faint">\uFF08\u7F3A\u7701 ', '</b> <span class="small faint">(default '],
    ['<button class="btn primary sm" data-ctx-save>\u4FDD\u5B58\u5206\u53C9</button>', '<button class="btn primary sm" data-ctx-save>Save fork</button>'],
    ['<details class="md-raw"><summary>\u539F\u59CB\u8F7D\u8377\uFF08JSON\uFF09</summary><pre class="mono md-raw-pre">', '<details class="md-raw"><summary>Raw payload (JSON)</summary><pre class="mono md-raw-pre">'],
    ['<div class="apply-dock"><button class="btn primary block" data-ed-apply disabled>Apply\uFF08\u65E0\u53D8\u66F4\uFF09</button></div>', '<div class="apply-dock"><button class="btn primary block" data-ed-apply disabled>Apply (no changes)</button></div>'],
    ['<div class="card"><div class="card-b"><div class="small faint">\u4E0A\u4E0B\u6587\u9762\u6682\u4E0D\u53EF\u8FBE\uFF08', '<div class="card"><div class="card-b"><div class="small faint">Context face unreachable ('],
    ['<div class="card"><div class="card-b"><div class="small faint">\u4F1A\u8BDD\u6E05\u5355\u6682\u4E0D\u53EF\u5F97', '<div class="card"><div class="card-b"><div class="small faint">Session list unavailable'],
    ['<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u4E3A\u7A7A\uFF08\u96F6\u5F85\u529E\u2014\u2014\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div></div></div>', '<div class="card"><div class="card-b"><div class="small faint">Inbox is empty (zero to-dos \u2014 explicit empty state).</div></div></div>'],
    ['<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u6295\u5F71\u7F3A\u5E2D\uFF08\u5FEB\u7167\u672A\u643A\u5E26 /v1/inbox \u805A\u5408\uFF09\u2014\u2014\u5F85\u4F60\u5904\u7406\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u88C5\u5047\u3002</div></div></div>', '<div class="card"><div class="card-b"><div class="small faint">Inbox projection absent (snapshot carries no /v1/inbox aggregate) \u2014 your to-do face is explicitly absent, never faked.</div></div></div>'],
    ['<div class="card"><div class="card-h"><span class="h-title">\u6982\u89C8</span></div><div class="card-b">', '<div class="card"><div class="card-h"><span class="h-title">Overview</span></div><div class="card-b">'],
    ['<div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587\uFF08ContextPack \u5951\u7EA6\u9762\uFF09</span><span class="h-right">', '<div class="card-h"><span class="h-title">Context (ContextPack contract face)</span><span class="h-right">'],
    ['<div class="card-h"><span class="h-title">\u5386\u53F2</span></div><div class="card-b">', '<div class="card-h"><span class="h-title">History</span></div><div class="card-b">'],
    ['<div class="card-h"><span class="h-title">\u8349\u7A3F</span></div><div class="card-b">', '<div class="card-h"><span class="h-title">Draft</span></div><div class="card-b">'],
    ['<div class="card-h"><span class="h-title">\u8349\u7A3F</span><span class="h-right">', '<div class="card-h"><span class="h-title">Draft</span><span class="h-right">'],
    ['<div class="chat-grid"><div><h2 class="sec-title">\u4F1A\u8BDD\u6E05\u5355</h2>', '<div class="chat-grid"><div><h2 class="sec-title">Session list</h2>'],
    ['<div class="dd-grid"><div data-dd-tree><div class="small faint">\u4ED3\u6811\u73B0\u8BFB\u4E2D\u2026</div></div>', '<div class="dd-grid"><div data-dd-tree><div class="small faint">Loading repo tree\u2026</div></div>'],
    ['<div class="diff-line ctx">\u2026 \u5171 ', '<div class="diff-line ctx">\u2026 total '],
    ['<div class="mr-sub small">\u76F8\u5173\u5BF9\u8C61 ', '<div class="mr-sub small">Related objects '],
    ['<div class="ob-empty">scheme \u6682\u4E0D\u53EF\u8FBE\uFF08kinds \u805A\u5408\u7F3A\u5E2D\uFF09\u3002</div>', '<div class="ob-empty">scheme unreachable (kinds aggregate absent).</div>'],
    ['<div class="ob-empty">\u4E8B\u4EF6\u5168\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08', '<div class="ob-empty">Event detail unreachable ('],
    ['<div class="ob-empty">\u4ED3\u6811\u4E3A\u7A7A\uFF08\u8BDA\u5B9E\u7A7A\u6001\uFF09\u3002</div>', '<div class="ob-empty">Repo tree is empty (honest empty state).</div>'],
    ['<div class="ob-empty">\u4ED3\u6811\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="ob-empty">Repo tree unreadable ('],
    ['<div class="ob-empty">\u65E0\u9002\u7528\u89C4\u5219\u2014\u2014\u6309 kind/scope/tag/\u6B63\u6587\u5F15\u7528\u56DB\u8DEF\u68C0\u67E5\u96F6\u547D\u4E2D\uFF08\u68C0\u67E5\u8303\u56F4=', '<div class="ob-empty">No applicable rules \u2014 zero hits across kind/scope/tag/content-reference checks (check scope ='],
    ['<div class="ob-empty">\u6682\u65E0\u6807\u7B7E\u2014\u2014\u5BF9\u8C61\u6B63\u6587\u672A\u58F0\u660E spec.metadata.tags\uFF08\u58F0\u660E\u7ECF Apply \u843D\u5E93\u540E\u6B64\u5904\u73B0\u8BFB\uFF09\u3002</div>', '<div class="ob-empty">No tags \u2014 object content declares no spec.metadata.tags (appears here once applied via Apply).</div>'],
    ['<div class="ob-empty">\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002</div>', '<div class="ob-empty">This document has no content.</div>'],
    ['<div class="ob-empty">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="ob-empty">Content unreadable ('],
    ['<div class="ob-empty">\u8BE5 kind \u65E0\u5B57\u6BB5\u58F0\u660E\uFF08spec \u7F3A\u5E2D\uFF09\u3002</div>', '<div class="ob-empty">This kind has no field declarations (no spec).</div>'],
    ['<div class="ob-empty">\u8F6C\u5F55\u9762\u672A\u88C5\u914D\u2014\u2014\u5BBF\u4E3B\u6CE8\u5165 window.__CC.chatFaces \u540E\u73B0\u8BFB\uFF1B', '<div class="ob-empty">Transcript face not mounted \u2014 reads live once the host injects window.__CC.chatFaces;'],
    ['<div class="ob-empty">\uFF08\u7A7A\u58F0\u660E\uFF09</div>', '<div class="ob-empty">(empty declaration)</div>'],
    ['<div class="report-foot small faint">\u6570\u636E\u53EF\u4FE1\u5EA6\uFF1A', '<div class="report-foot small faint">Data confidence:'],
    ['<div class="small faint" style="grid-column:1/-1">\u2026\u53E6\u6709 ', '<div class="small faint" style="grid-column:1/-1">\u2026and '],
    ['<div class="small faint" style="margin-bottom:10px">\u5FEB\u7167\u672A\u643A\u5E26 attention \u805A\u5408\u2014\u2014\u672C\u5361\u7531\u4E8B\u4EF6\u5F52\u6863\u884C\uFF08ApprovalNeeded/Escalate\uFF09\u6295\u5F71\u751F\u6210\u3002</div>', '<div class="small faint" style="margin-bottom:10px">Snapshot carries no attention aggregate \u2014 this card is projected from archived event rows (ApprovalNeeded/Escalate).</div>'],
    ['<div class="small faint" style="margin-bottom:10px">\u5FEB\u7167\u672A\u643A\u5E26 attention \u805A\u5408\uFF0C\u4E8B\u4EF6\u5F52\u6863\u4EA6\u65E0\u5BA1\u6279/\u5347\u7EA7\u884C\u2014\u2014\u672C\u8F6E\u62A5\u544A\u6309\u5728\u518C\u4E8B\u5B9E\u6536\u655B\u4E3A\u96F6\u547D\u4E2D\uFF08\u4E0D\u88C5\u5047\uFF09\u3002</div>', '<div class="small faint" style="margin-bottom:10px">No attention aggregate in snapshot and no approval/escalation rows in the archive \u2014 this report converges to zero hits on recorded facts (never faked).</div>'],
    ['<div class="small faint" style="margin-bottom:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A', '<div class="small faint" style="margin-bottom:6px">Draft source fieldManager: '],
    ['<div class="small faint" style="margin-top:6px">\u57FA\u51C6=r', '<div class="small faint" style="margin-top:6px">base=r'],
    ['<div class="small faint" style="margin-top:6px">\u65E0\u8986\u76D6\uFF08\u751F\u6548\u503C=ContextPolicy \u7F3A\u7701\u2014\u2014PUT \u540E\u6B64\u5904\u73B0\u8BFB\u8986\u76D6\u884C\uFF09\u3002</div>', '<div class="small faint" style="margin-top:6px">No override (effective value = ContextPolicy default \u2014 PUT and the override row appears here).</div>'],
    ['<div class="small faint" style="margin-top:6px">\u73B0\u884C\u8986\u76D6\uFF1Arevision ', '<div class="small faint" style="margin-top:6px">Current override: revision '],
    ['<div class="small faint" style="margin-top:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A', '<div class="small faint" style="margin-top:8px">Draft source fieldManager: '],
    ['<div class="small faint" style="margin-top:8px">\u8F6C\u5F55\u4E0E\u7EC8\u7AEF\u9762\u7531\u5BBF\u4E3B faces \u6CE8\u5165\uFF08__CC.chatFaces / __CC.termFaces\uFF09\uFF1B', '<div class="small faint" style="margin-top:8px">Transcript & terminal faces are host-injected (__CC.chatFaces / __CC.termFaces); '],
    ['<div class="small faint" style="padding:8px 14px">\u2026\u5171 ', '<div class="small faint" style="padding:8px 14px">\u2026total '],
    ['<div class="small faint">\u4F1A\u8BDD\u4E0D\u5728\u6D3B\u6027\u8D26\uFF08\u65E0\u5FC3\u8DF3\u884C\uFF09\u2014\u2014\u5951\u7EA6\u9762\u4ECD\u53EF\u67E5\u770B/\u7F16\u8F91\u672C\u4F1A\u8BDD\u4E0A\u4E0B\u6587\u8986\u76D6\uFF08\u4E0D\u5192\u5145\u5728\u518C\uFF09\u3002</div>', `<div class="small faint">Session not in the liveness ledger (no heartbeat row) \u2014 the contract face still shows/edits this session's context override (never pretends it is live).</div>`],
    ['<div class="small faint">\u4FDD\u5B58\u5931\u8D25\uFF08', '<div class="small faint">Save failed ('],
    ['<div class="small faint">\u5199\u9762\u5173\u65AD\uFF08--allow-apply \u672A\u5F00\u542F\uFF09\u2014\u2014\u7F16\u8F91\u5206\u53C9\u4E0D\u53EF\u7528\uFF08\u4E0E\u5FC3\u8DF3\u6CE8\u518C\u9762\u540C\u5F00\u5173\uFF0C\u4E0D\u5192\u5145\u53EF\u7F16\u8F91\uFF09\u3002</div>', '<div class="small faint">Writes closed (--allow-apply off) \u2014 fork editing unavailable (same switch as the heartbeat registration face; never pretends to be editable).</div>'],
    ['<div class="small faint">\u5206\u7247\u7F13\u5B58\u9762\u7F3A\u5E2D\uFF08contextidx \u672A\u66B4\u9732 Shards\u2014\u2014\u8BDA\u5B9E\u964D\u7EA7\uFF0C\u4E0D\u5192\u5145\u7A7A\u7F13\u5B58\uFF09\u3002</div>', '<div class="small faint">Shard cache face absent (contextidx exposes no Shards \u2014 honest degradation, never pretends an empty cache).</div>'],
    ['<div class="small faint">\u5951\u7EA6\u672A\u643A\u5E26\u6B65\u9AA4\u660E\u7EC6\u2014\u2014\u4E0D\u663E\u793A\u4E0D\u7F16\u9020\u3002</div>', '<div class="small faint">Contract carries no step details \u2014 hidden, never fabricated.</div>'],
    ['<div class="small faint">\u5BA1\u6279\u8BB0\u5F55\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="small faint">Approval records unreadable ('],
    ['<div class="small faint">\u5DF2\u4FDD\u5B58\uFF08revision ', '<div class="small faint">Saved (revision '],
    ['<div class="small faint">\u5DF2\u51B3\u8BB0\u5F55\u6682\u65E0\u5224\u5B9A\u884C\uFF08\u6279\u51C6/\u9A73\u56DE\u5373\u51B3\u5373\u89C1\uFF09\u3002</div>', '<div class="small faint">No decision rows yet (approvals/rejections appear immediately).</div>'],
    ['<div class="small faint">\u5DF2\u64A4\u9500\u2014\u2014\u5199\u524D\u5FEB\u7167\u5DF2\u6062\u590D\u3002</div>', '<div class="small faint">Reverted \u2014 pre-write snapshot restored.</div>'],
    ['<div class="small faint">\u64A4\u9500\u5931\u8D25\uFF08', '<div class="small faint">Revert failed ('],
    ['<div class="small faint">\u65E0\u751F\u6548 tag \u5206\u7247\uFF08tagSig \u672A\u88C5\u914D\u2014\u2014\u7F16\u8F91\u5206\u53C9\u540E\u9010 tagKey \u6295\u5F71 TagCtxShard\uFF09\u3002</div>', '<div class="small faint">No active tag shards (tagSig not assembled \u2014 per-tagKey TagCtxShard projection after a fork edit).</div>'],
    ['<div class="small faint">\u65E0\u7D22\u5F15\u5019\u9009\uFF08tagSig \u672A\u88C5\u914D\u6216\u7D22\u5F15\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u5192\u5145\u88C5\u914D\u5305\u5185\u5BB9\uFF09\u3002</div>', '<div class="small faint">No index candidates (tagSig not assembled or zero index hits \u2014 never fabricates pack content).</div>'],
    ['<div class="small faint">\u65E0\u8349\u7A3F\u2014\u2014\u7F16\u8F91\u300C\u539F\u6587\u300D\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08\u5199\u7A7F /v1/drafts\uFF0C\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\u3002</div>', '<div class="small faint">No draft \u2014 editing "Source" accumulates a draft (writes through /v1/drafts; never touches the official document, no events).</div>'],
    ['<div class="small faint">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08', '<div class="small faint">Content unreadable ('],
    ['<div class="small faint">\u6B63\u6587\u73B0\u8BFB\u4E2D\u2026</div>', '<div class="small faint">Loading content\u2026</div>'],
    ['<div class="small faint">\u89C4\u5219\u547D\u4E2D\u5224\u5B9A\u4E2D\u2026</div>', '<div class="small faint">Evaluating rule hits\u2026</div>'],
    ['<div class="small faint">\u8BE5 kind \u6682\u65E0\u5728\u518C\u5BF9\u8C61\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>', '<div class="small faint">No registered objects for this kind (explicit empty state).</div>'],
    ['<div class="small faint">\u8BE5 workflow \u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>', '<div class="small faint">No runs for this workflow (explicit empty state).</div>'],
    ['<div class="small faint">\u8BE5\u5BF9\u8C61\u6682\u65E0\u5F52\u6863\u4E8B\u4EF6\u884C\uFF08DocApplied/DocReview \u7B49\u2014\u2014\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>', '<div class="small faint">No archived event rows for this object (DocApplied/DocReview etc. \u2014 explicit empty state).</div>'],
    ['<div class="small faint">\u9ED8\u8BA4\u5C55\u793A\u6E32\u67D3\u540E\u7684\u6210\u54C1\uFF1B\u70B9\u300C\u539F\u6587\u300D\u67E5\u770B\u4E0E\u7F16\u8F91\u6E90\u6587\u4EF6\u3002\u7F16\u8F91\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08POST /v1/drafts\u2014\u2014\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\uFF1BApply \u552F\u4E00\u5165\u53E3=\u53F3\u680F\u5438\u5E95\u6309\u94AE\u3002</div>', '<div class="small faint">Shows the rendered result by default; click "Source" to view/edit the raw file. Editing accumulates a draft (POST /v1/drafts \u2014 never touches the official document, no events); the only Apply entry is the docked button in the right column.</div>'],
    ['<div class="w-foot small faint">\u6570\u636E\u9762\uFF1AGET /v1/snapshot \u57FA\u7EBF + /v1/stream patch \u589E\u91CF\uFF08WS \u65AD\u7EBF\u81EA\u52A8\u964D\u7EA7\u8F6E\u8BE2\uFF09\uFF1B', '<div class="w-foot small faint">Data plane: GET /v1/snapshot baseline + /v1/stream patch increments (auto-degrades to polling on WS loss); '],
    ['<div data-ap-ledger style="margin-top:20px"><div class="small faint">\u5DF2\u51B3\u8BB0\u5F55\u73B0\u8BFB\u4E2D\u2026</div></div>', '<div data-ap-ledger style="margin-top:20px"><div class="small faint">Loading decision ledger\u2026</div></div>'],
    ['<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">\u4E0A\u4E0B\u6587\u6CE8\u6C34\u4E2D\u2026\uFF08GET /v1/sessions/', '<div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">Hydrating context\u2026 (GET /v1/sessions/'],
    ['<div><h2 class="sec-title">\u8F6C\u5F55</h2><div class="card"><div class="card-b transcript">', '<div><h2 class="sec-title">Transcript</h2><div class="card"><div class="card-b transcript">'],
    ['<h2 class="sec-title" style="margin-top:14px">\u4E0A\u4E0B\u6587</h2>', '<h2 class="sec-title" style="margin-top:14px">Context</h2>'],
    ['<h2 class="sec-title">kinds \u6458\u8981</h2>', '<h2 class="sec-title">kinds summary</h2>'],
    ['<h2 class="sec-title">\u4F1A\u8BDD\u6E05\u5355</h2>', '<h2 class="sec-title">Session list</h2>'],
    ['<h2 class="sec-title">\u5165\u53E3</h2>', '<h2 class="sec-title">Entry points</h2>'],
    ['<h2 class="sec-title">\u5BF9\u8C61</h2><div class="card">', '<h2 class="sec-title">Object</h2><div class="card">'],
    ['<h2 class="sec-title">\u5DF2\u51B3\u8BB0\u5F55</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>', '<h2 class="sec-title">Decision ledger</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>'],
    ['<h2 class="sec-title">\u5F53\u524D\u5FAA\u73AF</h2>', '<h2 class="sec-title">Current loop</h2>'],
    ['<h2 class="sec-title">\u6536\u4EF6\u7BB1\uFF08\u5F85\u4F60\u5904\u7406\uFF09</h2>', '<h2 class="sec-title">Inbox (awaiting you)</h2>'],
    ['<h2 class="sec-title">\u6700\u8FD1\u76F8\u5173\u4E8B\u4EF6</h2><div class="card">', '<h2 class="sec-title">Recent related events</h2><div class="card">'],
    ['<h2 class="sec-title">\u6B65\u9AA4</h2><div class="card">', '<h2 class="sec-title">Steps</h2><div class="card">'],
    ['<h2 class="sec-title">\u8FD0\u884C\u8BB0\u5F55</h2>', '<h2 class="sec-title">Run history</h2>'],
    ['<h3 class="sec-title" style="margin:10px 0 6px">chunks\uFF08\u7D22\u5F15 TopK \u5019\u9009\uFF09</h3>', '<h3 class="sec-title" style="margin:10px 0 6px">chunks (index TopK candidates)</h3>'],
    ['<h3 class="sec-title" style="margin:10px 0 6px">\u76F8\u5173 Tag \u5206\u7247\uFF08TagCtxShard\uFF09</h3>', '<h3 class="sec-title" style="margin:10px 0 6px">Related tag shards (TagCtxShard)</h3>'],
    ['<input class="inp" data-ctx-note placeholder="\u5206\u53C9\u5907\u6CE8\uFF08\u53EF\u9009\uFF09">', '<input class="inp" data-ctx-note placeholder="Fork note (optional)">'],
    ['<input class="inp" data-ctx-tags placeholder="tag \u96C6\uFF1AtagType:tag \u9017\u53F7/\u7A7A\u683C\u5206\u9694\uFF08\u6BCF\u7C7B\u2264', '<input class="inp" data-ctx-tags placeholder="tag set: tagType:tag, comma/space separated (per type \u2264'],
    ['<input class="inp" id="docs-q" placeholder="\u6309 kind/scope/name \u8FC7\u6EE4\u2026" value="', '<input class="inp" id="docs-q" placeholder="Filter by kind/scope/name\u2026" value="'],
    ['<input class="inp" id="ev-q" placeholder="\u6309\u7C7B\u578B/\u5BF9\u8C61/\u6458\u8981\u8FC7\u6EE4\u2026" value="', '<input class="inp" id="ev-q" placeholder="Filter by type/object/summary\u2026" value="'],
    ['<section class="card obj-content"><div class="card-h"><span class="h-title">\u5185\u5BB9</span><span class="h-right">', '<section class="card obj-content"><div class="card-h"><span class="h-title">Content</span><span class="h-right">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587</span></div><div class="card-b">', '<section class="card"><div class="card-h"><span class="h-title">Context</span></div><div class="card-b">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u64CD\u4F5C</span></div><div class="card-b">', '<section class="card"><div class="card-h"><span class="h-title">Actions</span></div><div class="card-b">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u6807\u7B7E</span></div><div class="card-b" data-panel="tags">', '<section class="card"><div class="card-h"><span class="h-title">Tags</span></div><div class="card-b" data-panel="tags">'],
    ['<section class="card"><div class="card-h"><span class="h-title">\u9002\u7528\u89C4\u5219</span></div><div class="card-b" data-panel="rules">', '<section class="card"><div class="card-h"><span class="h-title">Applicable rules</span></div><div class="card-b" data-panel="rules">'],
    ['<span class="chip ok tiny" data-ed-draft-chip>\u65E0\u8349\u7A3F</span>', '<span class="chip ok tiny" data-ed-draft-chip>No draft</span>'],
    ['<span class="faint small">\u2026\u5171 ', '<span class="faint small">\u2026total '],
    ['<span class="faint small">\u672A\u88C5\u914D\uFF08tagSigSource=', '<span class="faint small">not assembled (tagSigSource='],
    ['<span class="faint">\u8981\u7D20\u5B57\u6BB5\u7F3A\u5E2D\u2014\u2014\u4E0D\u7F16\u9020</span>', '<span class="faint">Essential fields absent \u2014 never fabricated</span>'],
    ['<span class="faint">\uFF08\u7A7A\uFF09</span>', '<span class="faint">(empty)</span>'],
    ['<span class="mono faint small">\uFF08\u5D4C\u5957\u7ED3\u6784\u89C1\u300CYAML\u300D\uFF09</span>', '<span class="mono faint small">(nested structures in "YAML")</span>'],
    ['<span class="small faint" data-conn-label>\u8FDE\u63A5\u4E2D\u2026</span></div></aside>', '<span class="small faint" data-conn-label>Connecting\u2026</span></div></aside>'],
    ['<span class="small faint">\u5FAA\u73AF\u72B6\u6001\u6682\u4E0D\u53EF\u8FBE\uFF08', '<span class="small faint">Loop status unreachable ('],
    ['<span class="small faint">\u672C\u5468\u671F\u6682\u65E0\u5FAA\u73AF\u8BB0\u5F55\u3002\u70B9\u706B\u4E0B\u4E00\u8F6E\uFF1A<code class="mono">corectl loop</code></span></div></div></div>', '<span class="small faint">No loop records this cycle. Ignite the next round: <code class="mono">corectl loop</code></span></div></div></div>'],
    ['<span class="small faint">\u8BC1\u636E\uFF1A\u2014</span>', '<span class="small faint">Evidence: \u2014</span>'],
    ['<span><span class="logo-name">Core Console</span><br><span class="logo-sub">web \u91CD\u6784\u8D77\u70B9</span></span></div>', '<span><span class="logo-name">Core Console</span><br><span class="logo-sub">web rebuild baseline</span></span></div>'],
    ['<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="\u539F\u6587\u7F16\u8F91\u2014\u2014\u53D8\u66F4\u5373\u7D2F\u79EF\u8349\u7A3F"', '<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="Edit source \u2014 changes accumulate a draft"'],
    ["<th>\u65F6\u523B</th><th>\u5224\u5B9A</th><th>\u5BF9\u8C61</th><th>\u5224\u5B9A\u4E3B\u4F53</th><th>rv</th></tr></thead><tbody>", "<th>Time</th><th>Decision</th><th>Object</th><th>Decided by</th><th>rv</th></tr></thead><tbody>"],
    [">\u64A4\u9500\u4E0A\u6B21</button>", ">Undo last</button>"],
    ["Apply \u5931\u8D25\uFF08", "Apply failed ("],
    ["Apply \u6210\u529F\u2014\u2014\u6587\u6863\u5DF2\u843D\u5E93\uFF08rv=", "Apply succeeded \u2014 document persisted (rv="],
    ["Apply \u843D\u5E93 ", "Apply persisted "],
    ["Apply\uFF08\u65E0\u53D8\u66F4\uFF09", "Apply (no changes)"],
    ["B vs \u843D\u5E93\u503C ", "B vs persisted "],
    ["B\uFF08\u73B0\u8BFB\uFF0C\u4E0D\u7F13\u5B58\uFF09 \xB7 ", "B (read live, not cached) \xB7 "],
    ["ConnectorOps / WorkService / Provider / Ref \u56DB\u65CF\uFF08\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\uFF09", "ConnectorOps / WorkService / Provider / Ref families (same rules for docs and resources)"],
    ["attention \u805A\u5408\u96F6\u547D\u4E2D\uFF08\u7A7A\u6001\u663E\u5F0F\u2014\u2014\u4E0D\u5192\u5145\u6709\u98CE\u9669\uFF09\u3002", "attention aggregate zero hits (explicit empty state \u2014 never fakes risk)."],
    ["dry-run \u9884\u68C0\u672A\u8FC7\uFF08", "dry-run precheck failed ("],
    ["kinds \u2192 scope \u2192 \u5BF9\u8C61\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF1B\u5BF9\u8C61\u540D\u53EF\u70B9\u8FDB\u5DE5\u4F5C\u53F0\uFF09", "kinds \u2192 scope \u2192 objects (snapshot aggregate projection; object names click into the workbench)"],
    ["severity \u6392\u5E8F \xB7 \u8BC1\u636E\u53EF\u70B9 \xB7 \u6570\u636E\u53EF\u4FE1\u5EA6\u663E\u5F0F", "severity-sorted \xB7 clickable evidence \xB7 explicit data confidence"],
    ["takeBudget \u987B\u4E3A ", "takeBudget must be "],
    ["workflow \u951A ", "workflow anchor "],
    ["\u2014\u2014contextidx \u672A\u88C5\u914D\uFF08\u9632\u5FA1\u9762\uFF09", " \u2014 contextidx not assembled (defensive face)"],
    ["\u2014\u2014\u4E0D\u5192\u5145\u5728\u518C\u4F1A\u8BDD\u3002</div></div></div>", " \u2014 never pretends to be a registered session.</div></div></div>"],
    ["\u2014\u2014\u4E50\u89C2\u9501\u51B2\u7A81\uFF1A\u7A3F\u5DF2\u4FDD\u7559\uFF0C\u8BF7\u5BF9\u9F50\u6700\u65B0 rv \u540E\u91CD\u8BD5", " \u2014 optimistic-lock conflict: draft preserved; align to the latest rv and retry"],
    ["\u2014\u2014\u4E50\u89C2\u9501\u51B2\u7A81\uFF1A\u8BF7\u91CD\u8FDB\u9762\u677F\u5BF9\u9F50 revision \u540E\u91CD\u8BD5\uFF08\u7A3F\u4E0D\u8986\u76D6\uFF0C\u4E0D\u9759\u9ED8\u91CD\u57FA\uFF09", " \u2014 optimistic-lock conflict: reopen the panel to align revision and retry (draft not overwritten, no silent rebase)"],
    ["\u2014\u2014\u843D\u5E93\u5C06\u88AB\u62D2\uFF09", " \u2014 persistence would be rejected)"],
    ["\u2190 \u4EFB\u52A1\u4E2D\u5FC3", "\u2190 Task Center"],
    ["\u2190 \u4F1A\u8BDD", "\u2190 Sessions"],
    ["\u2190 \u5DE5\u4F5C\u6D41", "\u2190 Workflows"],
    ["\u25CF \u672A Apply\uFF08\u8349\u7A3F ", "\u25CF Not applied (draft "],
    ["\u300D\u540C\u7EC4", '" group'],
    ["\u4E00\u822C", "medium"],
    ["\u4E0D\u88C5\u5047\u2014\u2014scope \u8BCD\u5F62\u53EF\u80FD\u4E3A\u5B9E\u4F8B\u6876\uFF08\u6811\u5185\u53EF\u89C1\uFF09\u3002", "Never faked \u2014 scope word form may be an instance bucket (visible in the tree)."],
    ["\u4E25\u91CD", "critical"],
    ["\u4E50\u89C2\u9501\u5BF9\u9F50 r", "optimistic-lock align r"],
    ["\u4E50\u89C2\u9501\u6F02\u79FB\uFF08\u73B0 r", "optimistic-lock drift (now r"],
    ["\u4E8B\u4EF6", "Events"],
    ["\u4EFB\u52A1\u4E0D\u5728\u518C", "Task not registered"],
    ["\u4F1A\u8BDD ", "Session "],
    ["\u4F1A\u8BDD\u539F\u6587\u4E0D\u5165\u516C\u5171\u5E93\uFF08intake \u7559\u75D5\u9762\u4E3A\u51C6\uFF09\u3002</div>", "Raw session text stays out of the public library (intake ledger face is authoritative).</div>"],
    ["\u4F1A\u8BDD\u8BE6\u60C5", "Session detail"],
    ["\u4F1A\u8BDD\u8F6C\u5F55", "Session transcript"],
    ["\u4F5C\u7528\u57DF", "Scope"],
    ["\u4F9D\u8D56", "Depends on"],
    ["\u4F9D\u8D56 ", "Depends on "],
    ["\u516D\u7247\u56FA\u5B9A\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme", "Six fixed panels: tags/rules/context/actions/content/scheme"],
    ["\u5171 ", "total "],
    ["\u5173\u8054\u4F1A\u8BDD", "Related sessions"],
    ["\u5173\u8054\u8FD0\u884C", "Related runs"],
    ["\u5185\u7F6E\u670D\u52A1", "Built-in services"],
    ["\u51ED\u636E\u6E90", "Credential source"],
    ["\u539F\u6587", "Source"],
    ["\u539F\u6587\uFF08HEAD\uFF09", "Source (HEAD)"],
    ["\u53BB\u5904\u7406", "Handle"],
    ["\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "Go to Object Workbench"],
    ["\u53D7\u963B", "Blocked"],
    ["\u53EF\u64A4\u9500", "revertible"],
    ["\u540C\u540D\u5BF9\u8C61\u8DE8 kind/scope \u4E0D\u552F\u4E00\u2014\u2014\u8BF7\u4ECE\u7CFB\u7EDF\u6811\u8FDB\u5165\u3002", "Object name is ambiguous across kinds/scopes \u2014 enter via the System Tree."],
    ["\u540D\u79F0", "Name"],
    ["\u56DE Welcome", "Back to Welcome"],
    ["\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00", "Open in Docs"],
    ["\u57FA\u51C6=r", "base=r"],
    ["\u57FA\u51C6=r?", "base=r?"],
    ["\u57FA\u7840\u6C47\u603B", "Summary"],
    ["\u5916\u90E8\u4F9D\u8D56", "External dependencies"],
    ["\u5931\u8D25", "failed"],
    ["\u5BF9\u8BDD\u5165\u53E3\uFF08Agent\uFF09", "Chat entries (Agent)"],
    ["\u5BF9\u8C61\u5DE5\u4F5C\u53F0\uFF08\u516D\u7247\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme\uFF09", "Object Workbench (six panels: tags/rules/context/actions/content/scheme)"],
    ["\u5BF9\u8C61\u6B63\u6587\u5F15\u7528 ", "Object content reference "],
    ["\u5BF9\u8C61\u6B63\u6587\u6309 resourceVersion \u7F13\u5B58\uFF08rv \u672A\u53D8\u4E0D\u91CD\u8BFB\uFF09\u3002</div>", "Object content cached by resourceVersion (no re-read while rv unchanged).</div>"],
    ["\u5BF9\u8C61\u89E3\u6790\u6B67\u4E49\u6216\u7F3A\u5E2D", "Object resolution ambiguous or absent"],
    ["\u5DE5\u4F5C\u6D41\u4E0D\u5728\u518C", "Workflow not registered"],
    ["\u5DE5\u4F5C\u6D41\u8BE6\u60C5", "Workflow detail"],
    ["\u5DE6\u4FA7\u9009\u62E9\u4ED3\u6587\u4EF6\uFF08\u9ED8\u8BA4\u9996\u4ED3\u9996\u6587\u4EF6\uFF09\u3002", "Pick a repo file on the left (defaults to first repo, first file)."],
    ["\u5DF2\u5165\u5FEB\u7167", "in snapshot"],
    ["\u5DF2\u53D6\u6D88", "canceled"],
    ["\u5DF2\u5B8C\u6210", "done"],
    ["\u5DF2\u5E94\u7528\u7684\u7248\u672C", "Applied version"],
    ["\u5DF2\u6302\u8F7D\u8FDE\u63A5", "Mounted connectors"],
    ["\u5F53\u524D\u4F1A\u8BDD", "Current session"],
    ["\u5F53\u524D\u5FAA\u73AF", "Current loop"],
    ["\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF", "No loop in progress"],
    ["\u5F53\u524D\u8FC7\u6EE4\u96F6\u547D\u4E2D\u2014\u2014\u653E\u5BBD\u6761\u4EF6\u518D\u8BD5\u3002", "Filter has zero hits \u2014 relax the conditions and retry."],
    ["\u5F85\u542F\u52A8", "pending start"],
    ["\u5F85\u5B9A", "pending"],
    ["\u5F85\u5BA1\u533A=\u8349\u7A3F\u7F13\u5B58\uFF08DocReview \u7C7B\u540C\u6E90\uFF09\u222A ApprovalNeeded \u5F52\u6863\u884C\uFF08\u5DF2\u6279\u7684\u4E0D\u51FA\u73B0\uFF09\u3002\u6279\u51C6/\u9A73\u56DE\u5747=\u4EBA\u70B9\uFF1Bdiff=\u7A3F vs \u6587\u6863\u73B0\u503C\u73B0\u8BFB\u3002", "Pending = draft cache (DocReview same source) \u222A ApprovalNeeded archive rows (approved ones drop out). Approve/reject are human clicks; diff = draft vs live document, read fresh."],
    ["\u5F85\u5BA1\u533A\u4E3A\u7A7A\uFF08\u8349\u7A3F\u7F13\u5B58\u4E0E ApprovalNeeded \u5F52\u6863\u5747\u7A7A\uFF09", "Pending is empty (both draft cache and ApprovalNeeded archive are empty)"],
    ["\u5FAA\u73AF\u5524\u9192 #", "Loop wake #"],
    ["\u5FAA\u73AF\u72B6\u6001\u73B0\u8BFB\u4E2D\u2026", "Loading loop status\u2026"],
    ["\u5FEB\u7167", "Snapshot"],
    ["\u5FEB\u7167 events \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot events aggregate is empty (never faked)."],
    ["\u5FEB\u7167 objects \u805A\u5408\u4E3A\u7A7A\u6216\u8BE5 kind \u65E0\u5BF9\u8C61\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot objects aggregate empty or no objects for this kind (never faked)."],
    ["\u5FEB\u7167 proposals \u805A\u5408\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Snapshot proposals aggregate is empty (explicit empty state)."],
    ["\u5FEB\u7167 runs \u805A\u5408\u65E0\u6B64\u8FD0\u884C\uFF08\u53EF\u80FD\u5DF2\u51FA\u7A97\uFF09\u3002", "Snapshot runs aggregate has no such run (may have left the window)."],
    ["\u5FEB\u7167 workflows \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot workflows aggregate is empty (never faked)."],
    ["\u5FEB\u7167\u4EFB\u52A1\u6E05\u5355\u65E0\u6B64 id\uFF08\u53EF\u80FD\u5DF2\u5B8C\u6210\u51FA\u7A97\u6216 id \u6709\u8BEF\uFF09\u3002", "No such id in the snapshot task list (may be done and out of window, or a wrong id)."],
    ["\u5FEB\u7167\u6682\u65E0\u5BF9\u8C61\u679A\u4E3E", "Snapshot carries no object enumeration"],
    ["\u5FEB\u7167\u805A\u5408\u65E0\u6B64 workflow\uFF08\u53EF\u80FD\u5DF2\u4E0B\u7EBF\u6216\u540D\u6709\u8BEF\uFF09\u3002", "Snapshot aggregate has no such workflow (may be offline or a wrong name)."],
    ["\u6253\u5F00\u7CFB\u7EDF\u6811", "Open System Tree"],
    ["\u6279\u51C6\uFF08= Apply\uFF09", "Approve (= Apply)"],
    ["\u6279\u91CF\u6279\u51C6\u901A\u8DEF\u5F85\u540E\u7AEF\u8BCD\u5F62\u2014\u2014\u9010\u6761\u6279\u51C6\uFF08\u8BDA\u5B9E\u6CE8\u8BB0\uFF09", "Batch approval awaits a backend word form \u2014 approve one by one (honest note)"],
    ["\u6392\u961F\u4E2D", "queued"],
    ["\u63D0\u6848\u9762\uFF08\u5217\u8868+\u8981\u7D20\u73B0\u8BFB\uFF09\u2014\u2014\u6279\u51C6/\u9A73\u56DE\u8D70\u5BA1\u6279\u4E0E Apply\uFF08LLM \u65E0 Apply \u6743\uFF09", "Proposal face (list + live essentials) \u2014 approve/reject go through Approvals and Apply (the LLM has no Apply right)"],
    ["\u64A4\u9500\u6700\u8FD1\u4E00\u6B21\u4E0A\u4E0B\u6587\u5206\u53C9\uFF08\u6062\u590D\u670D\u52A1\u7AEF\u5199\u524D\u5FEB\u7167\uFF09\uFF1F", "Revert the latest context fork (restore the server-side pre-write snapshot)?"],
    ["\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\u2014\u2014\u540C\u4E00\u6E32\u67D3\u6838\uFF1B\u25CF=\u6709\u672A Apply \u8349\u7A3F", "Docs and resources share the rules \u2014 one render core; \u25CF = unapplied draft exists"],
    ["\u6587\u6863\u7F16\u8F91\u5668", "Document editor"],
    ["\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09", "Document absent (registered draft base \u2014 Apply = create path)"],
    ["\u65E0\u7A3F", "no draft"],
    ["\u65E0\u8349\u7A3F", "No draft"],
    ["\u65F6\u95F4\u672A\u77E5", "time unknown"],
    ["\u667A\u80FD\u5F52\u7EB3", "Smart summary"],
    ["\u6682\u65E0\u5728\u518C\u5DE5\u4F5C\u6D41", "No registered workflows"],
    ["\u6682\u65E0\u5F85\u529E", "No to-dos"],
    ["\u6682\u65E0\u63D0\u6848", "No proposals"],
    ["\u6700\u8FD1\u52A8\u4F5C", "Last action"],
    ["\u6700\u8FD1\u52A8\u4F5C\uFF1A\u2014", "Last action: \u2014"],
    ["\u6700\u8FD1\u9519\u8BEF", "Last error"],
    ["\u670D\u52A1\u5B9E\u4F8B", "Service instances"],
    ["\u672A\u5206\u7EA7", "unranked"],
    ["\u672A\u5206\u7EC4", "Ungrouped"],
    ["\u672A\u547D\u540D\u4E8B\u4EF6", "Unnamed event"],
    ["\u672A\u547D\u540D\u4EFB\u52A1", "Unnamed task"],
    ["\u672A\u547D\u540D\u5173\u6CE8\u9879", "Unnamed attention item"],
    ["\u672A\u547D\u540D\u63D0\u6848", "Unnamed proposal"],
    ["\u672A\u6CE8\u5165=\u6E05\u5355\u964D\u7EA7\u5FEB\u7167\u6295\u5F71\uFF0C\u8F6C\u5F55\u9762\u663E\u5F0F\u7F3A\u5E2D\u8BCD\u5F62\u3002</div>", "Not injected = list degrades to snapshot projection; the transcript face is explicitly absent.</div>"],
    ["\u672A\u6CE8\u518C\u58F0\u660E kind\u2014\u2014\u5168\u91CF\u679A\u4E3E\u515C\u5E95", "Unregistered declared kind \u2014 falls back to full enumeration"],
    ["\u672A\u6CE8\u660E\u751F\u6210\u65B9\u5F0F", "generation method unspecified"],
    ["\u672A\u77E5", "unknown"],
    ["\u672A\u77E5\u8584\u5C4F", "Unknown thin screen"],
    ["\u672A\u9009\u6863", "No file selected"],
    ["\u672C\u7A97\u53E3\u65E0\u5173\u6CE8\u9879", "No attention items in this window"],
    ["\u67E5\u770B ", "View "],
    ["\u6807\u7B7E\u300C", 'Tag "'],
    ["\u6807\u9898", "Title"],
    ["\u6B63\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08", "Content unreachable ("],
    ["\u6CE8\u6C34\u4E2D\u2026", "Hydrating\u2026"],
    ["\u7248\u672C\u53F7", "Version"],
    ["\u72B6\u6001", "Status"],
    ["\u73B0\u8BFB\u4E2D\u2026", "Loading\u2026"],
    ["\u767D\u540D\u5355\u4ED3\u6811\uFF08GET /v1/design-docs\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF09", "Whitelisted repo tree (GET /v1/design-docs \u2014 HEAD authoritative, zero cache)"],
    ["\u76EE\u6807", "Target"],
    ["\u76EE\u6807\u6587\u6863", "Target document"],
    ["\u76F8\u5173\u6587\u6863", "Related documents"],
    ["\u770B root \u4F1A\u8BDD", "View root session"],
    ["\u786E\u8BA4\u6279\u51C6\uFF08Apply\uFF09", "Confirm approve (Apply)"],
    ["\u7A3F: ", "Draft: "],
    ["\u7A97\u53E3\u5185\u65E0\u4E8B\u4EF6", "No events in window"],
    ["\u7A97\u53E3\u5185\u65E0\u6587\u6863", "No documents in window"],
    ["\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u3002", "Waiting for the first snapshot aggregate."],
    ["\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u6216 patch \u589E\u91CF", "Waiting for first snapshot aggregate or patch increment"],
    ["\u7ED3\u675F\u4E8E", "ended at"],
    ["\u7F51\u7EDC", "Network"],
    ["\u805A\u5408\u9762\uFF08\u5FEB\u7167 workflows/runs\uFF09\uFF1B\u8FD0\u884C\u72EC\u7ACB\u9875\u89C1\u8BE6\u60C5", "Aggregate face (snapshot workflows/runs); runs have their own detail page"],
    ["\u8349\u7A3F\u222A\u5BA1\u6279\u4E8B\u4EF6\u2014\u2014\u70B9\u51FB\u5904\u7406", "Drafts \u222A approval events \u2014 click to handle"],
    ["\u8349\u7A3F\u4FDD\u5B58\u5931\u8D25\uFF08", "Draft save failed ("],
    ["\u843D\u5E93\u4E2D\u2026", "Persisting\u2026"],
    ["\u89C4\u5219\u539F\u6587\u63D0\u53CA ", "Rule source mentions "],
    ["\u89C4\u5219\u6587\u6863\u65CF\u7D22\u5F15\u6682\u4E0D\u53EF\u8FBE\u2014\u2014\u547D\u4E2D\u5224\u5B9A\u6682\u7F3A\uFF08\u4E0D\u5192\u5145\u65E0\u9002\u7528\uFF09\u3002", "Rule doc family index unreachable \u2014 hit evaluation unavailable (never pretends none apply)."],
    ["\u89E6\u53D1\u4E8B\u4EF6", "Trigger event"],
    ["\u8BC1\u636E", "Evidence"],
    ["\u8BC1\u636E\u2014\u2014\u70B9\u51FB\u8FDB\u5165", "Evidence \u2014 click to open"],
    ["\u8BE5 scope \u6682\u65E0\u5728\u518C\u5BF9\u8C61", "No registered objects for this scope"],
    ["\u8BF4\u660E\u6587\u6863", "Guide"],
    ["\u8D1F\u8D23 ", "Owner "],
    ["\u8FC7\u6EE4\u96F6\u547D\u4E2D\u6216\u5FEB\u7167\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Filter zero hits or snapshot empty (explicit empty state)."],
    ["\u8FD0\u884C\u72B6\u6001", "Run status"],
    ["\u8FD0\u884C\u8BB0\u5F55\u4E0D\u5728\u518C", "Run record not registered"],
    ["\u8FD0\u884C\u8BE6\u60C5", "Run detail"],
    ["\u91CD\u8981", "important"],
    ["\u9519\u8BEF", "error"],
    ["\u9700\u6C42\u63D0\u6848", "Requirement proposal"],
    ["\u9884\u68C0\u4E2D\u2026", "Prechecking\u2026"],
    ["\u9A8C\u8BC1\u4E2D", "verifying"],
    ["\uFF08faces \u672A\u6CE8\u5165\u4E14\u5FEB\u7167 sessions \u7F3A\u5E2D\uFF09", "(faces not injected and snapshot sessions absent)"],
    ["\uFF08git \u6743\u5A01\u96F6\u7F13\u5B58\uFF09 \xB7 \u8349\u7A3F/Apply \u901A\u8DEF\u7ECF\u7F16\u8F91\u5668\u5C4F\uFF08doc/\u2026\uFF09\u3002</div>", "(git authoritative, zero cache) \xB7 draft/Apply path goes through the editor screen (doc/\u2026).</div>"],
    ["\uFF08rv \u672A\u53D8\u4E0D\u91CD\u8BFB\u6B63\u6587\u2014\u2014\u964D\u7EA7\u8F6E\u8BE2\u8DEF\u5F84\u540C\u6837\u751F\u6548\uFF09</span>", "(content not re-read while rv unchanged \u2014 applies to the degraded polling path too)</span>"],
    ["\uFF08\u4E50\u89C2\u9501\u57FA\u51C6\uFF09", "(optimistic-lock base)"],
    ["\uFF08\u4E50\u89C2\u9501\u671F\u671B\u503C\u2014\u2014409 \u4E0D\u9759\u9ED8\u91CD\u57FA\uFF09\xB7 \u7A3F ", "(optimistic-lock expected \u2014 409 never silently rebases) \xB7 draft "],
    ["\uFF08\u5BF9\u8BDD\u662F\u4E00\u7B49\u9762\u2014\u2014\u540C\u4E00 dsh \u4F1A\u8BDD\u673A\u5236\u6295\u5F71\uFF09", "(chat is a first-class face \u2014 projection of the same dsh session mechanism)"],
    ["\uFF08\u5BF9\u8C61\u4E0D\u5728\u518C\u2014\u2014apply \u5C06\u521B\u5EFA\u901A\u8DEF\uFF09", "(object not registered \u2014 apply is the create path)"],
    ["\uFF08\u65E0\u5DEE\u5F02\uFF09", "(no diff)"],
    ["\uFF08\u9762\u73B0\u503C\u6682\u4E0D\u53EF\u8BFB\uFF1A", "(face current value unreadable: "],
    ["\uFF09\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF0C\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>", ") \u2014 HEAD authoritative, zero cache; re-enter to read fresh.</div>"],
    ["\uFF09\u2014\u2014\u4E0D\u5192\u5145\u65E0\u5FAA\u73AF\u3002</span></div></div></div>", ") \u2014 never pretends there is no loop.</span></div></div></div>"],
    ["\uFF09\u2014\u2014\u5224\u5B9A\u7559\u75D5\u6682\u53EA=apply_log\u3002</div>", ") \u2014 decision trail is currently apply_log only.</div>"],
    ["\uFF09\u2014\u2014\u5DE5\u4F5C\u526F\u672C\u4ECD\u5728\u5185\u5B58\uFF0C\u9000\u51FA\u524D\u4E0D\u4E22", ") \u2014 the working copy stays in memory, not lost until you leave"],
    ["\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB", ") \u2014 re-enter after recovery to read fresh"],
    ["\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>", ") \u2014 re-enter after recovery to read fresh.</div>"],
    ["\uFF09\u2014\u2014\u6808\u7A7A\u4E0D\u53EF\u64A4\uFF08\u4E0D\u5192\u5145\u64A4\u9500\u6210\u529F\uFF09\u3002</div>", ") \u2014 empty stack cannot revert (never pretends a successful revert).</div>"],
    ["\uFF09\u2014\u2014\u7A3F\u4FDD\u7559\u8BED\u4E49\uFF0C\u672A\u843D\u5E93\u3002", ") \u2014 draft-preserved semantics, not persisted."],
    ["\uFF09\u2014\u2014\u91CD\u8FDB\u5373\u6062\u590D\u3002</div>", ") \u2014 re-enter to restore.</div>"],
    ["\uFF09\u2014\u2014\u96F6\u526F\u4F5C\u7528\uFF0C\u7A3F\u5DF2\u4FDD\u7559\u3002", ") \u2014 zero side effects, draft preserved."],
    ["\uFF09\u3002\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div></div></div>", "). Re-enter after recovery to read fresh.</div></div></div>"],
    ["\uFF09\uFF1F\u4E50\u89C2\u9501\u57FA\u51C6 r", ")? optimistic-lock base r"],
    ["\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>", ": none (explicit empty state).</div>"],
    ["\uFF1B\u8D85\u9650 tag \u6536\u655B ", "; over-limit tags converge "],
    ["\uFF1Fdry-run \u5DF2\u8FC7\uFF08rv=", "? dry-run passed (rv="],
    /* §19 全页可用性批——EN 回填（plain 缺译全量） */
    ["\u6587\u6863\u7F16\u8F91\u5668", "Document editor"],
    ["\u8FD0\u884C\u8BE6\u60C5", "Run detail"],
    ["Service \u5927\u5C4F", "Service dashboard"],
    ["\u7CFB\u7EDF\u8BE6\u60C5", "System detail"],
    ["\uFF08\u65E0\u5DEE\u5F02\uFF09", "(no diff)"],
    ["\u540D\u79F0", "Name"],
    ["\u4F5C\u7528\u57DF", "Scope"],
    ["\u7248\u672C\u53F7", "Version"],
    ["\u76EE\u6807\u6587\u6863", "Target doc"],
    ["\u8BF4\u660E\u6587\u6863", "Description"],
    ["\u76F8\u5173\u6587\u6863", "Related docs"],
    ["\u5916\u90E8\u4F9D\u8D56", "External deps"],
    ["\u670D\u52A1\u5B9E\u4F8B", "Service instances"],
    ["\u5185\u7F6E\u670D\u52A1", "Built-in services"],
    ["\u5BF9\u8BDD\u5165\u53E3\uFF08Agent\uFF09", "Dialog entry (Agent)"],
    ["\u8FD0\u884C\u72B6\u6001", "Run status"],
    ["\u5DF2\u5E94\u7528\u7684\u7248\u672C", "Applied version"],
    ["\u5DF2\u6302\u8F7D\u8FDE\u63A5", "Mounted connections"],
    ["\u51ED\u636E\u6E90", "Credential source"],
    ["\u5F53\u524D\u4F1A\u8BDD", "Current session"],
    ["\u6807\u9898", "Title"],
    ["\u4F9D\u8D56", "Deps"],
    ["\u539F\u6587", "Source"],
    ["\u67E5\u770B\u5BF9\u8C61", "View object"],
    ["\u5C55\u5F00", "Expand"],
    ["\u6211", "Me"],
    ["\u751F\u6210\u7ED3\u8BBA", "Conclusion"],
    ["\u5B8C\u6574\u6587\u6863\u6A21\u5F0F\uFF1A", "Full document: "],
    ["\u5B8C\u6574\u6587\u6863\u6A21\u5F0F", "Full document"],
    ["\u5199\u5165\u8349\u7A3F", "To draft"],
    ["\u601D\u8003\u8FC7\u7A0B", "Thinking"],
    ["\u5DE5\u5177\u8FC7\u7A0B", "Tool trace"],
    ["\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "Object workspace"],
    ["\u7A3F: ", "Draft: "],
    ["\uFF08\u5BF9\u8C61\u4E0D\u5728\u518C\u2014\u2014apply \u5C06\u521B\u5EFA\u901A\u8DEF\uFF09", "(object not registered \u2014 apply will create it)"],
    ["\u57FA\u51C6=r?", "base=r?"],
    ["\u57FA\u51C6=r", "base=r"],
    ["\uFF08\u9762\u73B0\u503C\u6682\u4E0D\u53EF\u8BFB\uFF1A", "(face value unreadable: "],
    [" \xB7 \u7A3F ", " \xB7 draft "],
    ["B vs \u843D\u5E93\u503C ", "B vs stored "],
    ["B\uFF08\u73B0\u8BFB\uFF0C\u4E0D\u7F13\u5B58\uFF09 \xB7 ", "B (live read, no cache) \xB7 "],
    ["\u9884\u68C0\u4E2D\u2026", "Dry-run\u2026"],
    ["\u6279\u51C6\uFF08= Apply\uFF09", "Approve (= Apply)"],
    ["dry-run \u9884\u68C0\u672A\u8FC7\uFF08", "dry-run precheck failed ("],
    ["\uFF09\u2014\u2014\u96F6\u526F\u4F5C\u7528\uFF0C\u7A3F\u5DF2\u4FDD\u7559\u3002", ") \u2014 zero side effects, draft kept."],
    ["\u4E50\u89C2\u9501\u5BF9\u9F50 r", "optimistic lock aligned r"],
    ["\u4E50\u89C2\u9501\u6F02\u79FB\uFF08\u73B0 r", "optimistic lock drift (now r"],
    [" vs \u7A3F\u57FA r", " vs draft base r"],
    ["\u2014\u2014\u843D\u5E93\u5C06\u88AB\u62D2\uFF09", " \u2014 apply would be rejected)"],
    ["\u786E\u8BA4\u6279\u51C6\uFF08Apply\uFF09", "Confirm approve (Apply) "],
    ["\uFF1Fdry-run \u5DF2\u8FC7\uFF08rv=", "? dry-run passed (rv="],
    ["\u843D\u5E93\u4E2D\u2026", "Applying\u2026"],
    ["Apply \u5931\u8D25\uFF08", "Apply failed ("],
    ["\uFF09\u2014\u2014\u7A3F\u4FDD\u7559\u8BED\u4E49\uFF0C\u672A\u843D\u5E93\u3002", ") \u2014 draft kept, not applied."],
    ["\u53BB\u5904\u7406", "Handle"],
    ["\u5F85\u5BA1\u533A=\u8349\u7A3F\u7F13\u5B58\uFF08DocReview \u7C7B\u540C\u6E90\uFF09\u222A ApprovalNeeded \u5F52\u6863\u884C\uFF08\u5DF2\u6279\u7684\u4E0D\u51FA\u73B0\uFF09\u3002\u6279\u51C6/\u9A73\u56DE\u5747=\u4EBA\u70B9\uFF1Bdiff=\u7A3F vs \u6587\u6863\u73B0\u503C\u73B0\u8BFB\u3002", "Pending = draft cache (DocReview same source) \u222A ApprovalNeeded archived rows (granted ones drop out). Approve/reject are human clicks; diff = draft vs live value."],
    ["\u6279\u91CF\u6279\u51C6\u901A\u8DEF\u5F85\u540E\u7AEF\u8BCD\u5F62\u2014\u2014\u9010\u6761\u6279\u51C6\uFF08\u8BDA\u5B9E\u6CE8\u8BB0\uFF09", "Batch approve awaits backend verb \u2014 approve one by one (honest note)"],
    ["\u5F85\u5BA1\u533A\u4E3A\u7A7A\uFF08\u8349\u7A3F\u7F13\u5B58\u4E0E ApprovalNeeded \u5F52\u6863\u5747\u7A7A\uFF09", "Nothing pending (draft cache and ApprovalNeeded archive both empty)"],
    ["\u56DE Welcome", "Back to Welcome"],
    ["\u2014 \u9009\u62E9 session\uFF08\u6216\u8F93\u5165\uFF09\u2014", "\u2014 pick a session (or type) \u2014"],
    ["\u6216\u624B\u586B session id", "or type a session id"],
    ["\u63D0\u4EA4\u9700\u6C42\uFF0F\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\u2014\u2014\u4F8B\u5982\uFF1A\u8FDB\u884C\u4EE3\u7801\u5BA1\u6838\uFF0C\u770B\u770B\u5B9E\u9645\u548C\u6587\u6863\u5DEE\u8DDD\u6709\u591A\u5927\uFF0C\u5E76\u8FDB\u884C\u4FEE\u590D", "Submit a request / find a doc / update a doc \u2014 e.g. run a code audit, measure the actual-vs-doc gap, and fix it"],
    ["\u53D1\u9001", "Send"],
    ["Enter \u53D1\u9001 \xB7 queue \u8BED\u4E49\uFF08\u9001\u8FBE agent \u4F1A\u8BDD\uFF1B\u7ED3\u8BBA\u7531 agent \u843D\u6587\u6863\u540E\u7ECF\u72B6\u6001\u5FBD\u4E0E\u8282\u70B9\u56FE\u53EF\u89C1\uFF09", "Enter to send \xB7 queue semantics (delivered to the agent session; the conclusion lands in docs by the agent, visible via status badge and node graph)"],
    ["\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u9001\u8FBE\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u5DF2\u53D1\u9001", "Dialog channel not injected (chatfaces.js absent) \u2014 delivery face explicitly absent, never fakes a send"],
    ["\u7EBF\u7A0B\u4E3A\u7A7A\u2014\u2014\u53D1\u7B2C\u4E00\u6761\u9700\u6C42\uFF1Bagent \u7684\u5B8C\u6574\u8F68\u8FF9\u5728\u539F\u751F\u754C\u9762\uFF08\u65B0\u7A97\u53E3\uFF09\u53EF\u89C1\u3002", "Thread empty \u2014 send the first request; the agent\u2019s full track is visible in the native UI (new window)."],
    ["\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08", "Body unreadable ("],
    ["\u5185\u5BB9\u7F3A\u5E2D", "content absent"],
    ["\u5173\u95ED", "Close"],
    ["\u5199\u5165\u8349\u7A3F\uFF08\u7F16\u8F91\u5668\uFF09", "To draft (editor)"],
    ["\u5DF2\u9001\u8FBE\uFF08queue\uFF09\u2014\u2014agent \u4F1A\u8BDD\u5DF2\u53D7\u7406\u3002\u7ED3\u8BBA\u5C06\u7531 agent \u751F\u6210\u5E76\u81EA\u884C\u843D\u6587\u6863\uFF1B\u6587\u6863\u72B6\u6001\u5FBD\u4E0E workflow \u8282\u70B9\u56FE\u968F\u540E\u53EF\u89C1\u3002\u539F\u751F\u8F68\u8FF9\uFF1A", "Delivered (queue) \u2014 agent session accepted. The conclusion will be produced by the agent and land in docs on its own; watch the doc status badge and workflow node graph. Native track:"],
    ["\u9001\u8FBE\u5931\u8D25\uFF08", "Delivery failed ("],
    ["\u5BBF\u4E3B\u539F\u751F\u754C\u9762\uFF08\u6743\u5A01\u8F68\u8FF9\uFF09", "Host native UI (authoritative track)"],
    ["rootAgent\u2014\u2014\u627E\u6587\u6863\u3001\u66F4\u65B0\u6587\u6863\uFF1Bref \u53EF\u70B9\uFF0C\u7ED3\u8BBA\u53EF\u5168\u6587\u6863\u89C2\u770B", "rootAgent \u2014 find and update docs; refs are clickable, conclusions viewable as full documents"],
    ["\u65B0\u7A97\u53E3\u539F\u751F\u754C\u9762", "Native UI in new window"],
    ["\u5BF9\u8BDD\u7EBF\u7A0B", "Dialog thread"],
    ["\u672C\u5730\u7EBF\u7A0B\u53EA\u8BB0\u672C\u63A7\u5236\u53F0\u53D1\u51FA\u7684\u6D88\u606F\uFF1B\u6743\u5A01\u8F68\u8FF9=\u5BBF\u4E3B\u539F\u751F\u4F1A\u8BDD\uFF08\u4E0B\u65B9 expand \u5185\u5D4C\uFF09\u3002", "Local thread only records messages sent from this console; the authoritative track is the host native session (embedded below)."],
    ["\u4E0A\u4E0B\u6587\u6CE8\u6C34\u4E2D\u2026\uFF08GET /v1/sessions/", "Hydrating context\u2026 (GET /v1/sessions/"],
    ["\u65E0\u7A3F", "No draft"],
    ["\u6B63\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08", "Body unreachable ("],
    ["\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB", ") \u2014 reopen after recovery to read live"],
    ["scheme\uFF08frontmatter \u58F0\u660E\u2014\u2014expand \u67E5\u770B\u5B8C\u6574\uFF09", "scheme (frontmatter declaration \u2014 expand for full)"],
    ["\u25CF \u672A Apply\uFF08\u8349\u7A3F ", "\u25CF unapplied (draft "],
    [" \u5904\u53D8\u66F4)", " change(s))"],
    ["\u65E0\u8349\u7A3F", "No draft"],
    ["Apply\uFF08\u65E0\u53D8\u66F4\uFF09", "Apply (no changes)"],
    ["\u6D41\u7A0B\uFF1A\u5199\u4FEE\u6539\u610F\u56FE \u2192 AI \u51FA\u5168\u6587\u5EFA\u8BAE \u2192 \u5199\u5165\u8349\u7A3F \u2192 \u53F3\u4E0B Apply\uFF08\u4E50\u89C2\u9501+\u9884\u68C0\u7167\u8D70\u2014\u2014AI \u65E0\u7279\u6743\uFF09\u3002", "Flow: write the change intent \u2192 AI proposes a full draft \u2192 to draft \u2192 Apply at bottom-right (optimistic lock + precheck as usual \u2014 AI has no privilege)."],
    ["\u4FEE\u6539\u610F\u56FE\uFF0C\u5982\uFF1A\u8865\u5145 X \u8282\uFF1B\u4FEE\u6B63 Y \u5904\u53E3\u5F84\u2026", "Change intent, e.g. add section X; fix wording at Y\u2026"],
    ["\u8BF7\u6C42 AI \u5EFA\u8BAE", "Ask AI for a proposal"],
    ["\u5EFA\u8BAE\u5C06\u73B0\u8BFB\u4E8E\u6B64\u2014\u2014\u5199\u5165\u8349\u7A3F\u540E\u8D70\u65E2\u6709 Apply \u901A\u8DEF\u3002", "The proposal will appear here \u2014 send to draft, then use the existing Apply path."],
    ["AI \u901A\u9053\u672A\u6CE8\u5165\uFF08__CC.chatFaces.docFix \u7F3A\u5E2D\uFF09\u2014\u2014\u53EF\u5728\u300C\u539F\u6587\u300D\u89C6\u56FE\u624B\u7F16\uFF0C\u6216\u7ECF\u4F1A\u8BDD\u5C4F\u5BF9\u8BDD\u540E\u4EBA\u5DE5\u8D34\u56DE\u3002", "AI channel not injected (__CC.chatFaces.docFix absent) \u2014 edit manually in Source view, or dialog on the session screen and paste back."],
    ["AI \u73B0\u8BFB\u4E2D\u2026", "AI working\u2026"],
    ["\u7A7A\u5EFA\u8BAE", "Empty proposal"],
    ["AI \u5EFA\u8BAE\u5DF2\u5230\uFF08", "AI proposal arrived ("],
    ["\u5DF2\u5199\u5165\u8349\u7A3F\u2014\u2014\u53F3\u4E0B Apply \u843D\u5E93\uFF08\u4E50\u89C2\u9501\u7167\u8D70\uFF09\u3002", "Written to draft \u2014 Apply at bottom-right (optimistic lock as usual)."],
    ["AI \u5EFA\u8BAE\u5931\u8D25\uFF08", "AI proposal failed ("],
    ["\uFF08\u4E50\u89C2\u9501\u671F\u671B\u503C\u2014\u2014409 \u4E0D\u9759\u9ED8\u91CD\u57FA\uFF09\xB7 \u7A3F ", "(optimistic-lock expected \u2014 409 never silently rebases) \xB7 draft "],
    ["\u8349\u7A3F\u4FDD\u5B58\u5931\u8D25\uFF08", "Draft save failed ("],
    ["\uFF09\u2014\u2014\u5DE5\u4F5C\u526F\u672C\u4ECD\u5728\u5185\u5B58\uFF0C\u9000\u51FA\u524D\u4E0D\u4E22", ") \u2014 working copy stays in memory, not lost before exit"],
    ["\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09", "Doc absent (registered draft base \u2014 Apply = creation path)"],
    ["\uFF08\u4E50\u89C2\u9501\u57FA\u51C6\uFF09", "(optimistic-lock base)"],
    ["\u6CE8\u6C34\u4E2D\u2026", "Hydrating\u2026"],
    ["Apply \u843D\u5E93 ", "Apply stored "],
    ["\uFF09\uFF1F\u4E50\u89C2\u9501\u57FA\u51C6 r ", ")? optimistic-lock base r "],
    ["\u2014\u2014\u4E50\u89C2\u9501\u51B2\u7A81\uFF1A\u7A3F\u5DF2\u4FDD\u7559\uFF0C\u8BF7\u5BF9\u9F50\u6700\u65B0 rv \u540E\u91CD\u8BD5", " \u2014 optimistic-lock conflict: draft kept, align to latest rv and retry"],
    ["Apply \u6210\u529F\u2014\u2014\u6587\u6863\u5DF2\u843D\u5E93\uFF08rv=", "Applied \u2014 doc stored (rv="],
    ["\u5BF9\u8C61\u89E3\u6790\u6B67\u4E49\u6216\u7F3A\u5E2D", "Object resolution ambiguous or absent"],
    ["\u540C\u540D\u5BF9\u8C61\u8DE8 kind/scope \u4E0D\u552F\u4E00\u2014\u2014\u8BF7\u4ECE\u7CFB\u7EDF\u6811\u8FDB\u5165\u3002", "Same-named object across kinds/scopes \u2014 enter via the system tree."],
    ["\u6253\u5F00\u7CFB\u7EDF\u6811", "Open system tree"],
    ["\u67E5\u770B ", "View "],
    ["workflow \u951A ", "workflow anchor "],
    [" \u2192 \u5DE5\u4F5C\u6D41", " \u2192 workflow"],
    ["\u5173\u8054\u8FD0\u884C", "Related runs"],
    ["\u4F1A\u8BDD ", "Session "],
    ["\u4F1A\u8BDD\u8BE6\u60C5", "Session detail"],
    ["\u5FAA\u73AF\u5524\u9192 #", "Loop wake #"],
    ["\u672A\u547D\u540D\u4E8B\u4EF6", "Unnamed event"],
    ["\u7A97\u53E3\u5185\u65E0\u4E8B\u4EF6", "No events in window"],
    ["\u5F53\u524D\u8FC7\u6EE4\u96F6\u547D\u4E2D\u2014\u2014\u653E\u5BBD\u6761\u4EF6\u518D\u8BD5\u3002", "Filter has zero hits \u2014 loosen it and retry."],
    ["\u5FEB\u7167 events \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot events aggregate is empty (no faking)."],
    ["\u72B6\u6001", "Status"],
    ["\u73B0\u8BFB\u4E2D\u2026", "Reading live\u2026"],
    ["kinds \u2192 scope \u2192 \u5BF9\u8C61\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF1B\u5BF9\u8C61\u540D\u53EF\u70B9\u8FDB\u5DE5\u4F5C\u53F0\uFF09", "kinds \u2192 scope \u2192 objects (snapshot projection; names click into the workspace)"],
    ["\u5FEB\u7167\u6682\u65E0\u5BF9\u8C61\u679A\u4E3E", "Snapshot has no object enumeration"],
    ["\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u3002", "Waiting for the first snapshot aggregate."],
    ["\u76EE\u5F55\u6811\u5F0F\u6587\u6863\u6E05\u5355\u2014\u2014tag \u8FC7\u6EE4+\u7248\u672C/\u91CD\u8981\u5EA6\u5217\uFF1B\u70B9\u884C\u8FDB\u6E32\u67D3\u89C6\u56FE\u4E0E AI \u4FEE\u6587\u6863", "Tree-style doc list \u2014 tag filter + version/importance columns; click a row for rendered view and AI doc-fix"],
    ["\u672A\u6CE8\u518C\u58F0\u660E kind\u2014\u2014\u5168\u91CF\u679A\u4E3E\u515C\u5E95", "No declared kinds registered \u2014 full enumeration fallback"],
    ["\u91CD\u8981\u5EA6\u2014", "importance \u2014"],
    ["\u6253\u5F00", "Open"],
    ["\u6E32\u67D3\u89C6\u56FE", "Rendered view"],
    ["\u7A97\u53E3\u5185\u65E0\u6587\u6863", "No docs in window"],
    ["\u8FC7\u6EE4\u96F6\u547D\u4E2D\u6216\u5FEB\u7167\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Filter zero-hit or snapshot empty (explicit empty state)."],
    ["tag/\u91CD\u8981\u5EA6\u6765\u81EA\u7F13\u5B58\u6B63\u6587\u7684 frontmatter\uFF08\u9884\u53D6\u7A97\u53E3 24 \u7BC7\u2014\u2014\u672A\u7F13\u5B58\u7BC7\u663E\u793A \u2014\uFF0C\u73B0\u5F00\u5373\u8BFB\uFF09\u3002", "Tags/importance come from cached body frontmatter (prefetch window 24 \u2014 uncached show \u2014, read live on open)."],
    ["\u7F51\u7EDC", "Network"],
    ["\u539F\u6587\uFF08HEAD\uFF09", "Source (HEAD)"],
    ["\u767D\u540D\u5355\u4ED3\u6811\uFF08GET /v1/design-docs\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF09", "Whitelisted repo tree (GET /v1/design-docs \u2014 HEAD authoritative, zero cache)"],
    ["\u672A\u9009\u6863", "No file selected"],
    ["\u5DE6\u4FA7\u9009\u62E9\u4ED3\u6587\u4EF6\uFF08\u9ED8\u8BA4\u9996\u4ED3\u9996\u6587\u4EF6\uFF09\u3002", "Pick a repo file on the left (defaults to first repo, first file)."],
    ["ConnectorOps / WorkService / Provider / Ref \u56DB\u65CF\uFF08\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\uFF09", "Four families: ConnectorOps / WorkService / Provider / Ref (docs and resources, same rules)"],
    ["\u9700\u6C42\u63D0\u6848", "Proposals"],
    ["\u63D0\u6848\u9762\uFF08\u5217\u8868+\u8981\u7D20\u73B0\u8BFB\uFF09\u2014\u2014\u6279\u51C6/\u9A73\u56DE\u8D70\u5BA1\u6279\u4E0E Apply\uFF08LLM \u65E0 Apply \u6743\uFF09", "Proposal face (list + live elements) \u2014 approve/reject go through approvals and Apply (LLM has no Apply right)"],
    ["\u672A\u547D\u540D\u63D0\u6848", "Unnamed proposal"],
    ["\u5F85\u5B9A", "Pending"],
    ["\u6682\u65E0\u63D0\u6848", "No proposals"],
    ["\u5FEB\u7167 proposals \u805A\u5408\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002", "Snapshot proposals aggregate is empty (explicit empty state)."],
    [" \u7684\u5BF9\u8C61\u4E0E\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF09", "\u2019s objects and events (snapshot projection)"],
    ["\u8BE5 scope \u6682\u65E0\u5728\u518C\u5BF9\u8C61", "No registered objects under this scope"],
    ["\u4E0D\u88C5\u5047\u2014\u2014scope \u8BCD\u5F62\u53EF\u80FD\u4E3A\u5B9E\u4F8B\u6876\uFF08\u6811\u5185\u53EF\u89C1\uFF09\u3002", "No faking \u2014 the scope word may be an instance bucket (visible in the tree)."],
    ["\u672A\u77E5\u8584\u5C4F", "Unknown thin screen"],
    ["\u4E1A\u52A1\u63A7\u5236\u53F0", "Business console"],
    ["\u4E1A\u52A1\u89C6\u89D2\u5360\u4F4D\u2014\u2014\u540E\u7EED ServiceAgent \u767B\u5F55\u540E\u5728\u6B64\u627F\u8F7D\u4E1A\u52A1\u5927\u5C4F\u4E0E\u64CD\u4F5C\u9762\uFF08\u5F53\u524D\u96F6\u5185\u5BB9\uFF0C\u663E\u5F0F\u5360\u4F4D\u4E0D\u88C5\u5047\uFF09\u3002", "Business-view placeholder \u2014 a later ServiceAgent login will host the business dashboard here (zero content now; explicit placeholder, no faking)."],
    ["\u672A\u5206\u7EC4", "Ungrouped"],
    ["\u89C4\u5219\u6587\u6863\u65CF\u7D22\u5F15\u6682\u4E0D\u53EF\u8FBE\u2014\u2014\u547D\u4E2D\u5224\u5B9A\u6682\u7F3A\uFF08\u4E0D\u5192\u5145\u65E0\u9002\u7528\uFF09\u3002", "Rule-doc family index unreachable \u2014 hit decision missing (never fakes \u201Cno applicable rules\u201D)."],
    ["\u6807\u7B7E\u300C", "Tag \u201C"],
    ["\u300D\u540C\u7EC4", "\u201D same group"],
    ["\u89C4\u5219\u539F\u6587\u63D0\u53CA ", "Rule text mentions "],
    ["\u5BF9\u8C61\u6B63\u6587\u5F15\u7528 ", "Object body references "],
    ["\u5DF2\u5165\u5FEB\u7167", "In snapshot"],
    ["\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00", "Open in doc library"],
    [" \u7F3A", " missing"],
    ["\u5BF9\u8C61\u5DE5\u4F5C\u53F0\uFF08\u516D\u7247\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme\uFF09", "Object workspace (six panes: tags/rules/context/actions/content/scheme)"],
    ["\u5FEB\u7167", "Snapshot"],
    ["\u516D\u7247\u56FA\u5B9A\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme", "Six fixed panes: tags/rules/context/actions/content/scheme"],
    ["\u5FEB\u7167 objects \u805A\u5408\u4E3A\u7A7A\u6216\u8BE5 kind \u65E0\u5BF9\u8C61\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot objects aggregate empty or no objects of this kind (no faking)."],
    ["takeBudget \u987B\u4E3A ", "takeBudget must be "],
    [" \u7684\u6570\u503C\uFF08ContextPolicy \u9884\u7B97\u57DF\uFF09", " (ContextPolicy budget domain)"],
    [" \u6B21", " time(s)"],
    [" \xB7 \u6700\u8FD1 ", " \xB7 recent "],
    ["\u65F6\u95F4\u672A\u77E5", "time unknown"],
    [" \u603B\u2264", " total \u2264"],
    ["\u53EF\u64A4\u9500", "revertible"],
    [" \xB7 \u5907\u6CE8 ", " \xB7 note "],
    [" \xB7 \u57DF ", " \xB7 domain "],
    ["\u5171 ", "total "],
    [" \u7247 \xB7 stale ", " shard(s) \xB7 stale "],
    [" \xB7 \u91CD\u5EFA\u9608\u503C ", " \xB7 rebuild threshold "],
    ["\u2014\u2014\u4E50\u89C2\u9501\u51B2\u7A81\uFF1A\u8BF7\u91CD\u8FDB\u9762\u677F\u5BF9\u9F50 revision \u540E\u91CD\u8BD5\uFF08\u7A3F\u4E0D\u8986\u76D6\uFF0C\u4E0D\u9759\u9ED8\u91CD\u57FA\uFF09", " \u2014 optimistic-lock conflict: reopen the panel, align revision, retry (draft not overwritten, never silently rebased)"],
    ["\uFF1B\u8D85\u9650 tag \u6536\u655B ", "; over-limit tags converged "],
    [" \u4E2A\uFF1A", ": "],
    ["\u64A4\u9500\u6700\u8FD1\u4E00\u6B21\u4E0A\u4E0B\u6587\u5206\u53C9\uFF08\u6062\u590D\u670D\u52A1\u7AEF\u5199\u524D\u5FEB\u7167\uFF09\uFF1F", "Revert the last context fork (restore the server-side pre-write snapshot)?"],
    ["\u2014\u2014contextidx \u672A\u88C5\u914D\uFF08\u9632\u5FA1\u9762\uFF09", " \u2014 contextidx not assembled (defensive face)"],
    ["\u9A8C\u8BC1\u4E2D", "Verifying"],
    ["\u6392\u961F\u4E2D", "Queued"],
    ["\u5F85\u542F\u52A8", "Pending start"],
    ["\u53D7\u963B", "Blocked"],
    ["\u5931\u8D25", "Failed"],
    ["\u5DF2\u5B8C\u6210", "Done"],
    ["\u5DF2\u53D6\u6D88", "Canceled"],
    ["\u672A\u547D\u540D\u4EFB\u52A1", "Unnamed task"],
    ["\u4F9D\u8D56 ", "Deps "],
    ["\u8D1F\u8D23 ", "Owner "],
    ["\u6700\u8FD1\u52A8\u4F5C\uFF1A\u2014", "Last action: \u2014"],
    ["\u5173\u8054\u4F1A\u8BDD", "Related session"],
    [" \u9879", " item(s)"],
    ["\u2190 \u4EFB\u52A1\u4E2D\u5FC3", "\u2190 Task center"],
    ["\u4EFB\u52A1\u4E0D\u5728\u518C", "Task not registered"],
    ["\u5FEB\u7167\u4EFB\u52A1\u6E05\u5355\u65E0\u6B64 id\uFF08\u53EF\u80FD\u5DF2\u5B8C\u6210\u51FA\u7A97\u6216 id \u6709\u8BEF\uFF09\u3002", "No such id in the snapshot task list (completed out of window, or wrong id)."],
    ["\u6700\u8FD1\u52A8\u4F5C", "Last action"],
    ["\u6700\u8FD1\u9519\u8BEF", "Last error"],
    ["\u4E25\u91CD", "Critical"],
    ["\u91CD\u8981", "Major"],
    ["\u4E00\u822C", "Minor"],
    ["\u672A\u5206\u7EA7", "Ungraded"],
    ["\u8BC1\u636E\u2014\u2014\u70B9\u51FB\u8FDB\u5165", "Evidence \u2014 click through"],
    ["\u8BC1\u636E", "Evidence"],
    ["\u672A\u547D\u540D\u5173\u6CE8\u9879", "Unnamed attention item"],
    ["\u667A\u80FD\u5F52\u7EB3", "Smart digest"],
    ["\u57FA\u7840\u6C47\u603B", "Basic summary"],
    ["\u672A\u6CE8\u660E\u751F\u6210\u65B9\u5F0F", "Generation method unstated"],
    [" \xB7 \u68C0\u67E5\u8303\u56F4 ", " \xB7 check scope "],
    [" \xB7 \u622A\u81F3 ", " \xB7 as of "],
    [" \xB7 \u7F3A\u5E2D\u6570\u636E\uFF1A", " \xB7 absent data: "],
    ["severity \u6392\u5E8F \xB7 \u8BC1\u636E\u53EF\u70B9 \xB7 \u6570\u636E\u53EF\u4FE1\u5EA6\u663E\u5F0F", "severity-sorted \xB7 clickable evidence \xB7 data trust stated"],
    ["\u672C\u7A97\u53E3\u65E0\u5173\u6CE8\u9879", "No attention items in this window"],
    ["attention \u805A\u5408\u96F6\u547D\u4E2D\uFF08\u7A7A\u6001\u663E\u5F0F\u2014\u2014\u4E0D\u5192\u5145\u6709\u98CE\u9669\uFF09\u3002", "attention aggregate zero-hit (explicit empty \u2014 never fakes risk)."],
    ["\u5F53\u524D\u5FAA\u73AF", "Current loop"],
    ["\u89E6\u53D1\u4E8B\u4EF6", "Fire event"],
    ["\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF", "No loop in progress"],
    ["\u5168\u90E8", "All"],
    ["\u7A3F", "Draft"],
    ["\u5347\u7EA7", "Escalate"],
    ["\u65E0\u4F1A\u8BDD", "No session"],
    ["\u4E8B\u4EF6", "Events"],
    ["\u7A3F\u222A\u5F85\u5BA1\u4E8B\u4EF6", "drafts \u222A pending events"],
    ["\u6253\u5F00\u5BF9\u8BDD\u7EBF\u7A0B", "Open dialog thread"],
    ["\u6B63\u5728\u8DD1\u7684 Workflows", "Running workflows"],
    ["\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279", "Approvals needed recently"],
    ["\u6838\u5FC3 Agent", "Core agents"],
    ["\u603B\u89C8", "Overview"],
    ["\u7CFB\u7EDF\u89C6\u89D2\u5927\u5C4F\u2014\u2014\u6B63\u5728\u8DD1\u7684\u4EFB\u52A1\u3001\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279\u3001\u6838\u5FC3 Agent \u4E00\u773C\u53EF\u8BFB", "System-view dashboard \u2014 running tasks, approvals needed, core agents at a glance"],
    ["Service", "Service"],
    ["Service \u89C6\u89D2\u5927\u5C4F\uFF08\u4E0E\u7CFB\u7EDF\u5927\u5C4F\u540C\u6784\u5E03\u5C40\u2014\u2014\u540E\u7EED ServiceAgent \u767B\u5F55\u5373\u843D\u6B64\u89C6\u56FE\uFF09", "Service-view dashboard (same layout as the system one \u2014 a later ServiceAgent login lands here)"],
    ["ServiceDesc \u5728\u518C\u2014\u2014agents=", "ServiceDesc registered \u2014 agents="],
    ["ServiceDesc \u5BF9\u8C61\u4E0D\u5728\u518C\uFF08\u8BCD\u5F62\u6295\u5F71\u81EA\u5FEB\u7167\u2014\u2014\u4E0D\u5192\u5145\u5728\u518C\uFF09", "ServiceDesc object not registered (projected from snapshot \u2014 never fakes registration)"],
    ["\u8FDB\u884C\u4E2D", "Running"],
    ["\u5355\u8282\u70B9\uFF08\u58F0\u660E\u672A\u6295\u5F71\u76F8\u4F4D\uFF09", "Single node (declared phases not projected)"],
    ["\u70B9\u51FB\u8282\u70B9\u770B\u8BE5\u76F8\u4F4D\u7684\u4E8B\u4EF6\u94FE\u4E0E\u6D89\u53CA\u6587\u6863\uFF1B\u63A8\u8FDB\u7531 patch \u589E\u91CF\u5B9E\u65F6\u5237\u65B0\u3002", "Click a node for that phase\u2019s event chain and involved docs; progress refreshes live via patches."],
    ["\u7A97\u53E3\u5185\u65E0\u8BE5\u76F8\u4F4D\u4E8B\u4EF6\uFF08\u8BDA\u5B9E\u7A7A\u6001\uFF09\u3002", "No events for this phase in window (honest empty)."],
    ["\u6D89\u53CA", "Involved"],
    ["\u4E8B\u4EF6\u94FE\uFF08\u5F52\u6863\u7A97\u53E3\uFF09\uFF1A", "Event chain (archive window):"],
    ["\u672A\u77E5", "Unknown"],
    [" \u6B21\u8FD0\u884C", " run(s)"],
    ["\u805A\u5408\u9762\uFF08\u5FEB\u7167 workflows/runs\uFF09\uFF1B\u8FD0\u884C\u72EC\u7ACB\u9875\u89C1\u8BE6\u60C5", "Aggregate face (snapshot workflows/runs); runs have their own detail pages"],
    ["\u6682\u65E0\u5728\u518C\u5DE5\u4F5C\u6D41", "No registered workflows"],
    ["\u5FEB\u7167 workflows \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002", "Snapshot workflows aggregate is empty (no faking)."],
    ["\u5DE5\u4F5C\u6D41\u8BE6\u60C5", "Workflow detail"],
    ["\u2190 \u5DE5\u4F5C\u6D41", "\u2190 Workflows"],
    ["\u5DE5\u4F5C\u6D41\u4E0D\u5728\u518C", "Workflow not registered"],
    ["\u5FEB\u7167\u805A\u5408\u65E0\u6B64 workflow\uFF08\u53EF\u80FD\u5DF2\u4E0B\u7EBF\u6216\u540D\u6709\u8BEF\uFF09\u3002", "No such workflow in the snapshot aggregate (offline, or wrong name)."],
    ["\u76EE\u6807", "Target"],
    ["\u8FD0\u884C\u8BB0\u5F55\u4E0D\u5728\u518C", "Run not registered"],
    ["\u5FEB\u7167 runs \u805A\u5408\u65E0\u6B64\u8FD0\u884C\uFF08\u53EF\u80FD\u5DF2\u51FA\u7A97\uFF09\u3002", "No such run in the snapshot runs aggregate (out of window)."],
    ["\u7B49\u5BA1\u6279", "Awaiting approval"],
    ["\u8FD0\u884C\u4E2D", "Running"],
    ["\u53BB\u5BA1\u6279 \u2192", "To approvals \u2192"],
    ["\u5F53\u524D\u76F8\u4F4D", "Current phase"],
    ["\u4E8B\u4EF6\u65F6\u95F4\u7EBF", "Event timeline"],
    ["\u5355\u8282\u70B9", "Single node"],
    ["\u6B63\u5728\u53D1\u751F\u2014\u2014\u63A8\u8FDB\u7531 patch \u5B9E\u65F6\u5237\u65B0", "Happening now \u2014 progress refreshes live via patches"],
    ["\u5386\u53F2\u76F8\u4F4D\u2014\u2014\u4E8B\u4EF6\u6309\u5F52\u6863\u7A97\u53E3\u56DE\u653E", "Past phase \u2014 events replayed from the archive window"],
    ["\u53BB\u5904\u7406\u8FD9\u6761\u5BA1\u6279", "Handle this approval"],
    ["\u5F52\u6863\u7A97\u53E3\u5185\u6682\u65E0\u4E8B\u4EF6\u2014\u2014workflow \u63A8\u8FDB\u65F6\u6B64\u5904\u5B9E\u65F6\u843D\u884C\u3002", "No events in the archive window yet \u2014 rows land here live as the workflow advances."]
  ]);
  var DICTS = { en };
  var subs = /* @__PURE__ */ new Set();
  function t(s) {
    if (locale === "zh") return String(s);
    const d = DICTS[locale];
    return d && d.get(String(s)) || String(s);
  }
  function getLocale() {
    return locale;
  }
  function setLocale(l) {
    if (l !== "zh" && !DICTS[l]) return;
    locale = l;
    try {
      if (typeof localStorage !== "undefined") localStorage.setItem(KEY, l);
    } catch {
    }
    subs.forEach((f) => {
      try {
        f(l);
      } catch {
      }
    });
  }
  function onLocaleChange(f) {
    subs.add(f);
    return () => subs.delete(f);
  }

  // src/render/refs.js
  var REF_KINDS = "Rule|Goal|Desc|Agent|ServiceDesc|ServiceInstance|WorkService|ConnectorOps|Provider|Ref|Role|Group|Workflow|Task|Proposal|Notification|DesignDoc|BugDoc|Template|Event";
  function refRoute(kind, scope, name) {
    const k = encodeURIComponent(kind);
    const s = encodeURIComponent(scope || "global");
    const n = encodeURIComponent(name);
    return "obj/" + k + "/" + s + "/" + n;
  }
  function linkifyRefs(text2, names) {
    const s = String(text2 || "");
    const nameSet = names instanceof Set ? names : Array.isArray(names) ? new Set(names) : null;
    const re = new RegExp("(" + REF_KINDS + ")\\/([A-Za-z0-9_.\\/-]+?)\\/([A-Za-z0-9_.\\-\u4E00-\u9FFF]+)", "g");
    const connRe = /\b(connector|provider):([A-Za-z0-9_.-]+)/g;
    let html3 = "";
    let last = 0;
    const spans = [];
    let m;
    while ((m = re.exec(s)) !== null) {
      spans.push({ i: m.index, len: m[0].length, kind: m[1], scope: m[2], name: m[3] });
    }
    while ((m = connRe.exec(s)) !== null) {
      spans.push({ i: m.index, len: m[0].length, conn: m[1], cname: m[2] });
    }
    spans.sort((a, b) => a.i - b.i);
    for (const sp of spans) {
      if (sp.i < last) continue;
      html3 += esc(s.slice(last, sp.i));
      if (sp.conn) {
        html3 += '<span class="ref-chip" data-go="connectors">' + esc(sp.conn + ":" + sp.cname) + "</span>";
      } else {
        const route = refRoute(sp.kind, sp.scope, sp.name);
        html3 += '<span class="ref-chip" data-go="' + esc(route) + '" title="' + esc(t("\u67E5\u770B\u5BF9\u8C61")) + '">' + esc(sp.kind + "/" + sp.scope + "/" + sp.name) + "</span>";
      }
      last = sp.i + sp.len;
    }
    html3 += esc(s.slice(last));
    if (nameSet && nameSet.size) {
      for (const nm of nameSet) {
        const needle = esc(nm);
        if (!needle || !html3.includes(needle)) continue;
        html3 = html3.split(needle).join(
          '<span class="ref-chip" data-go="' + esc(refRoute("Doc", "global", nm)) + '" data-ref-name="' + esc(nm) + '">' + needle + "</span>"
        );
      }
    }
    return html3;
  }
  function expandBlock(innerHtml, label, opts) {
    const o = opts || {};
    return '<details class="ex-block' + (o.cls ? " " + esc(o.cls) : "") + (o.open ? " open" : "") + '"' + (o.open ? " open" : "") + '><summary class="ex-sum"><span class="ex-chev">\u25B8</span>' + esc(label || t("\u5C55\u5F00")) + '</summary><div class="ex-body">' + (innerHtml || "") + "</div></details>";
  }
  function renderMessage(msg, names) {
    const m = msg || {};
    const role = String(m.role || "assistant");
    const isUser = role === "user";
    const text2 = String(m.text || "");
    const html3 = '<div class="msg ' + (isUser ? "msg-user" : "msg-ai") + '"><div class="msg-avatar">' + (isUser ? "\u{1F464}" : "\u25C6") + '</div><div class="msg-body"><div class="msg-head"><b class="msg-name">' + esc(isUser ? t("\u6211") : String(m.name || "rootAgent")) + "</b>" + (m.seq != null ? '<span class="small faint mono">seq ' + esc(String(m.seq)) + "</span>" : "") + (m.ts ? '<span class="small faint">' + esc(m.ts) + "</span>" : "") + "</div>" + (m.kind === "conclusion" ? '<div class="concl-card"><div class="concl-t">' + esc(t("\u751F\u6210\u7ED3\u8BBA")) + '</div><div class="concl-text">' + linkifyRefs(text2, names) + "</div>" + (Array.isArray(m.refs) && m.refs.length ? '<div class="concl-refs">' + m.refs.map((r) => {
      const parts = String(r).split("/");
      return parts.length >= 3 ? '<span class="ref-chip" data-doc-full="' + esc(r) + '">' + esc(t("\u5B8C\u6574\u6587\u6863\u6A21\u5F0F\uFF1A")) + esc(parts[2]) + "</span>" : "";
    }).join("") : "") + '<div class="concl-acts"><button class="btn sm" data-doc-full="' + esc(String(m.refs && m.refs[0] || "")) + '">' + esc(t("\u5B8C\u6574\u6587\u6863\u6A21\u5F0F")) + '</button><button class="btn sm" data-go="chat/core-headless">' + esc(t("\u5199\u5165\u8349\u7A3F")) + "</button></div></div>" : '<div class="msg-text">' + linkifyRefs(text2, names) + "</div>") + (m.thinking ? expandBlock('<div class="msg-text faint">' + linkifyRefs(m.thinking, names) + "</div>", t("\u601D\u8003\u8FC7\u7A0B")) : "") + (m.tool ? expandBlock('<div class="msg-text mono faint">' + esc(String(m.tool)) + "</div>", t("\u5DE5\u5177\u8FC7\u7A0B")) : "") + "</div></div>";
    return html3;
  }

  // node_modules/marked/lib/marked.esm.js
  function _getDefaults() {
    return {
      async: false,
      breaks: false,
      extensions: null,
      gfm: true,
      hooks: null,
      pedantic: false,
      renderer: null,
      silent: false,
      tokenizer: null,
      walkTokens: null
    };
  }
  var _defaults = _getDefaults();
  function changeDefaults(newDefaults) {
    _defaults = newDefaults;
  }
  var noopTest = { exec: () => null };
  function edit(regex, opt = "") {
    let source = typeof regex === "string" ? regex : regex.source;
    const obj = {
      replace: (name, val) => {
        let valSource = typeof val === "string" ? val : val.source;
        valSource = valSource.replace(other.caret, "$1");
        source = source.replace(name, valSource);
        return obj;
      },
      getRegex: () => {
        return new RegExp(source, opt);
      }
    };
    return obj;
  }
  var other = {
    codeRemoveIndent: /^(?: {1,4}| {0,3}\t)/gm,
    outputLinkReplace: /\\([\[\]])/g,
    indentCodeCompensation: /^(\s+)(?:```)/,
    beginningSpace: /^\s+/,
    endingHash: /#$/,
    startingSpaceChar: /^ /,
    endingSpaceChar: / $/,
    nonSpaceChar: /[^ ]/,
    newLineCharGlobal: /\n/g,
    tabCharGlobal: /\t/g,
    multipleSpaceGlobal: /\s+/g,
    blankLine: /^[ \t]*$/,
    doubleBlankLine: /\n[ \t]*\n[ \t]*$/,
    blockquoteStart: /^ {0,3}>/,
    blockquoteSetextReplace: /\n {0,3}((?:=+|-+) *)(?=\n|$)/g,
    blockquoteSetextReplace2: /^ {0,3}>[ \t]?/gm,
    listReplaceTabs: /^\t+/,
    listReplaceNesting: /^ {1,4}(?=( {4})*[^ ])/g,
    listIsTask: /^\[[ xX]\] /,
    listReplaceTask: /^\[[ xX]\] +/,
    anyLine: /\n.*\n/,
    hrefBrackets: /^<(.*)>$/,
    tableDelimiter: /[:|]/,
    tableAlignChars: /^\||\| *$/g,
    tableRowBlankLine: /\n[ \t]*$/,
    tableAlignRight: /^ *-+: *$/,
    tableAlignCenter: /^ *:-+: *$/,
    tableAlignLeft: /^ *:-+ *$/,
    startATag: /^<a /i,
    endATag: /^<\/a>/i,
    startPreScriptTag: /^<(pre|code|kbd|script)(\s|>)/i,
    endPreScriptTag: /^<\/(pre|code|kbd|script)(\s|>)/i,
    startAngleBracket: /^</,
    endAngleBracket: />$/,
    pedanticHrefTitle: /^([^'"]*[^\s])\s+(['"])(.*)\2/,
    unicodeAlphaNumeric: /[\p{L}\p{N}]/u,
    escapeTest: /[&<>"']/,
    escapeReplace: /[&<>"']/g,
    escapeTestNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/,
    escapeReplaceNoEncode: /[<>"']|&(?!(#\d{1,7}|#[Xx][a-fA-F0-9]{1,6}|\w+);)/g,
    unescapeTest: /&(#(?:\d+)|(?:#x[0-9A-Fa-f]+)|(?:\w+));?/ig,
    caret: /(^|[^\[])\^/g,
    percentDecode: /%25/g,
    findPipe: /\|/g,
    splitPipe: / \|/,
    slashPipe: /\\\|/g,
    carriageReturn: /\r\n|\r/g,
    spaceLine: /^ +$/gm,
    notSpaceStart: /^\S*/,
    endingNewline: /\n$/,
    listItemRegex: (bull) => new RegExp(`^( {0,3}${bull})((?:[	 ][^\\n]*)?(?:\\n|$))`),
    nextBulletRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}(?:[*+-]|\\d{1,9}[.)])((?:[ 	][^\\n]*)?(?:\\n|$))`),
    hrRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}((?:- *){3,}|(?:_ *){3,}|(?:\\* *){3,})(?:\\n+|$)`),
    fencesBeginRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}(?:\`\`\`|~~~)`),
    headingBeginRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}#`),
    htmlBeginRegex: (indent) => new RegExp(`^ {0,${Math.min(3, indent - 1)}}<(?:[a-z].*>|!--)`, "i")
  };
  var newline = /^(?:[ \t]*(?:\n|$))+/;
  var blockCode = /^((?: {4}| {0,3}\t)[^\n]+(?:\n(?:[ \t]*(?:\n|$))*)?)+/;
  var fences = /^ {0,3}(`{3,}(?=[^`\n]*(?:\n|$))|~{3,})([^\n]*)(?:\n|$)(?:|([\s\S]*?)(?:\n|$))(?: {0,3}\1[~`]* *(?=\n|$)|$)/;
  var hr = /^ {0,3}((?:-[\t ]*){3,}|(?:_[ \t]*){3,}|(?:\*[ \t]*){3,})(?:\n+|$)/;
  var heading = /^ {0,3}(#{1,6})(?=\s|$)(.*)(?:\n+|$)/;
  var bullet = /(?:[*+-]|\d{1,9}[.)])/;
  var lheadingCore = /^(?!bull |blockCode|fences|blockquote|heading|html|table)((?:.|\n(?!\s*?\n|bull |blockCode|fences|blockquote|heading|html|table))+?)\n {0,3}(=+|-+) *(?:\n+|$)/;
  var lheading = edit(lheadingCore).replace(/bull/g, bullet).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/\|table/g, "").getRegex();
  var lheadingGfm = edit(lheadingCore).replace(/bull/g, bullet).replace(/blockCode/g, /(?: {4}| {0,3}\t)/).replace(/fences/g, / {0,3}(?:`{3,}|~{3,})/).replace(/blockquote/g, / {0,3}>/).replace(/heading/g, / {0,3}#{1,6}/).replace(/html/g, / {0,3}<[^\n>]+>\n/).replace(/table/g, / {0,3}\|?(?:[:\- ]*\|)+[\:\- ]*\n/).getRegex();
  var _paragraph = /^([^\n]+(?:\n(?!hr|heading|lheading|blockquote|fences|list|html|table| +\n)[^\n]+)*)/;
  var blockText = /^[^\n]+/;
  var _blockLabel = /(?!\s*\])(?:\\.|[^\[\]\\])+/;
  var def = edit(/^ {0,3}\[(label)\]: *(?:\n[ \t]*)?([^<\s][^\s]*|<.*?>)(?:(?: +(?:\n[ \t]*)?| *\n[ \t]*)(title))? *(?:\n+|$)/).replace("label", _blockLabel).replace("title", /(?:"(?:\\"?|[^"\\])*"|'[^'\n]*(?:\n[^'\n]+)*\n?'|\([^()]*\))/).getRegex();
  var list = edit(/^( {0,3}bull)([ \t][^\n]+?)?(?:\n|$)/).replace(/bull/g, bullet).getRegex();
  var _tag = "address|article|aside|base|basefont|blockquote|body|caption|center|col|colgroup|dd|details|dialog|dir|div|dl|dt|fieldset|figcaption|figure|footer|form|frame|frameset|h[1-6]|head|header|hr|html|iframe|legend|li|link|main|menu|menuitem|meta|nav|noframes|ol|optgroup|option|p|param|search|section|summary|table|tbody|td|tfoot|th|thead|title|tr|track|ul";
  var _comment = /<!--(?:-?>|[\s\S]*?(?:-->|$))/;
  var html = edit(
    "^ {0,3}(?:<(script|pre|style|textarea)[\\s>][\\s\\S]*?(?:</\\1>[^\\n]*\\n+|$)|comment[^\\n]*(\\n+|$)|<\\?[\\s\\S]*?(?:\\?>\\n*|$)|<![A-Z][\\s\\S]*?(?:>\\n*|$)|<!\\[CDATA\\[[\\s\\S]*?(?:\\]\\]>\\n*|$)|</?(tag)(?: +|\\n|/?>)[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|<(?!script|pre|style|textarea)([a-z][\\w-]*)(?:attribute)*? */?>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$)|</(?!script|pre|style|textarea)[a-z][\\w-]*\\s*>(?=[ \\t]*(?:\\n|$))[\\s\\S]*?(?:(?:\\n[ 	]*)+\\n|$))",
    "i"
  ).replace("comment", _comment).replace("tag", _tag).replace("attribute", / +[a-zA-Z:_][\w.:-]*(?: *= *"[^"\n]*"| *= *'[^'\n]*'| *= *[^\s"'=<>`]+)?/).getRegex();
  var paragraph = edit(_paragraph).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("|table", "").replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex();
  var blockquote = edit(/^( {0,3}> ?(paragraph|[^\n]*)(?:\n|$))+/).replace("paragraph", paragraph).getRegex();
  var blockNormal = {
    blockquote,
    code: blockCode,
    def,
    fences,
    heading,
    hr,
    html,
    lheading,
    list,
    newline,
    paragraph,
    table: noopTest,
    text: blockText
  };
  var gfmTable = edit(
    "^ *([^\\n ].*)\\n {0,3}((?:\\| *)?:?-+:? *(?:\\| *:?-+:? *)*(?:\\| *)?)(?:\\n((?:(?! *\\n|hr|heading|blockquote|code|fences|list|html).*(?:\\n|$))*)\\n*|$)"
  ).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("blockquote", " {0,3}>").replace("code", "(?: {4}| {0,3}	)[^\\n]").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex();
  var blockGfm = {
    ...blockNormal,
    lheading: lheadingGfm,
    table: gfmTable,
    paragraph: edit(_paragraph).replace("hr", hr).replace("heading", " {0,3}#{1,6}(?:\\s|$)").replace("|lheading", "").replace("table", gfmTable).replace("blockquote", " {0,3}>").replace("fences", " {0,3}(?:`{3,}(?=[^`\\n]*\\n)|~{3,})[^\\n]*\\n").replace("list", " {0,3}(?:[*+-]|1[.)]) ").replace("html", "</?(?:tag)(?: +|\\n|/?>)|<(?:script|pre|style|textarea|!--)").replace("tag", _tag).getRegex()
  };
  var blockPedantic = {
    ...blockNormal,
    html: edit(
      `^ *(?:comment *(?:\\n|\\s*$)|<(tag)[\\s\\S]+?</\\1> *(?:\\n{2,}|\\s*$)|<tag(?:"[^"]*"|'[^']*'|\\s[^'"/>\\s]*)*?/?> *(?:\\n{2,}|\\s*$))`
    ).replace("comment", _comment).replace(/tag/g, "(?!(?:a|em|strong|small|s|cite|q|dfn|abbr|data|time|code|var|samp|kbd|sub|sup|i|b|u|mark|ruby|rt|rp|bdi|bdo|span|br|wbr|ins|del|img)\\b)\\w+(?!:|[^\\w\\s@]*@)\\b").getRegex(),
    def: /^ *\[([^\]]+)\]: *<?([^\s>]+)>?(?: +(["(][^\n]+[")]))? *(?:\n+|$)/,
    heading: /^(#{1,6})(.*)(?:\n+|$)/,
    fences: noopTest,
    // fences not supported
    lheading: /^(.+?)\n {0,3}(=+|-+) *(?:\n+|$)/,
    paragraph: edit(_paragraph).replace("hr", hr).replace("heading", " *#{1,6} *[^\n]").replace("lheading", lheading).replace("|table", "").replace("blockquote", " {0,3}>").replace("|fences", "").replace("|list", "").replace("|html", "").replace("|tag", "").getRegex()
  };
  var escape = /^\\([!"#$%&'()*+,\-./:;<=>?@\[\]\\^_`{|}~])/;
  var inlineCode = /^(`+)([^`]|[^`][\s\S]*?[^`])\1(?!`)/;
  var br = /^( {2,}|\\)\n(?!\s*$)/;
  var inlineText = /^(`+|[^`])(?:(?= {2,}\n)|[\s\S]*?(?:(?=[\\<!\[`*_]|\b_|$)|[^ ](?= {2,}\n)))/;
  var _punctuation = /[\p{P}\p{S}]/u;
  var _punctuationOrSpace = /[\s\p{P}\p{S}]/u;
  var _notPunctuationOrSpace = /[^\s\p{P}\p{S}]/u;
  var punctuation = edit(/^((?![*_])punctSpace)/, "u").replace(/punctSpace/g, _punctuationOrSpace).getRegex();
  var _punctuationGfmStrongEm = /(?!~)[\p{P}\p{S}]/u;
  var _punctuationOrSpaceGfmStrongEm = /(?!~)[\s\p{P}\p{S}]/u;
  var _notPunctuationOrSpaceGfmStrongEm = /(?:[^\s\p{P}\p{S}]|~)/u;
  var blockSkip = /\[[^[\]]*?\]\((?:\\.|[^\\\(\)]|\((?:\\.|[^\\\(\)])*\))*\)|`[^`]*?`|<[^<>]*?>/g;
  var emStrongLDelimCore = /^(?:\*+(?:((?!\*)punct)|[^\s*]))|^_+(?:((?!_)punct)|([^\s_]))/;
  var emStrongLDelim = edit(emStrongLDelimCore, "u").replace(/punct/g, _punctuation).getRegex();
  var emStrongLDelimGfm = edit(emStrongLDelimCore, "u").replace(/punct/g, _punctuationGfmStrongEm).getRegex();
  var emStrongRDelimAstCore = "^[^_*]*?__[^_*]*?\\*[^_*]*?(?=__)|[^*]+(?=[^*])|(?!\\*)punct(\\*+)(?=[\\s]|$)|notPunctSpace(\\*+)(?!\\*)(?=punctSpace|$)|(?!\\*)punctSpace(\\*+)(?=notPunctSpace)|[\\s](\\*+)(?!\\*)(?=punct)|(?!\\*)punct(\\*+)(?!\\*)(?=punct)|notPunctSpace(\\*+)(?=notPunctSpace)";
  var emStrongRDelimAst = edit(emStrongRDelimAstCore, "gu").replace(/notPunctSpace/g, _notPunctuationOrSpace).replace(/punctSpace/g, _punctuationOrSpace).replace(/punct/g, _punctuation).getRegex();
  var emStrongRDelimAstGfm = edit(emStrongRDelimAstCore, "gu").replace(/notPunctSpace/g, _notPunctuationOrSpaceGfmStrongEm).replace(/punctSpace/g, _punctuationOrSpaceGfmStrongEm).replace(/punct/g, _punctuationGfmStrongEm).getRegex();
  var emStrongRDelimUnd = edit(
    "^[^_*]*?\\*\\*[^_*]*?_[^_*]*?(?=\\*\\*)|[^_]+(?=[^_])|(?!_)punct(_+)(?=[\\s]|$)|notPunctSpace(_+)(?!_)(?=punctSpace|$)|(?!_)punctSpace(_+)(?=notPunctSpace)|[\\s](_+)(?!_)(?=punct)|(?!_)punct(_+)(?!_)(?=punct)",
    "gu"
  ).replace(/notPunctSpace/g, _notPunctuationOrSpace).replace(/punctSpace/g, _punctuationOrSpace).replace(/punct/g, _punctuation).getRegex();
  var anyPunctuation = edit(/\\(punct)/, "gu").replace(/punct/g, _punctuation).getRegex();
  var autolink = edit(/^<(scheme:[^\s\x00-\x1f<>]*|email)>/).replace("scheme", /[a-zA-Z][a-zA-Z0-9+.-]{1,31}/).replace("email", /[a-zA-Z0-9.!#$%&'*+/=?^_`{|}~-]+(@)[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+(?![-_])/).getRegex();
  var _inlineComment = edit(_comment).replace("(?:-->|$)", "-->").getRegex();
  var tag = edit(
    "^comment|^</[a-zA-Z][\\w:-]*\\s*>|^<[a-zA-Z][\\w-]*(?:attribute)*?\\s*/?>|^<\\?[\\s\\S]*?\\?>|^<![a-zA-Z]+\\s[\\s\\S]*?>|^<!\\[CDATA\\[[\\s\\S]*?\\]\\]>"
  ).replace("comment", _inlineComment).replace("attribute", /\s+[a-zA-Z:_][\w.:-]*(?:\s*=\s*"[^"]*"|\s*=\s*'[^']*'|\s*=\s*[^\s"'=<>`]+)?/).getRegex();
  var _inlineLabel = /(?:\[(?:\\.|[^\[\]\\])*\]|\\.|`[^`]*`|[^\[\]\\`])*?/;
  var link = edit(/^!?\[(label)\]\(\s*(href)(?:(?:[ \t]*(?:\n[ \t]*)?)(title))?\s*\)/).replace("label", _inlineLabel).replace("href", /<(?:\\.|[^\n<>\\])+>|[^ \t\n\x00-\x1f]*/).replace("title", /"(?:\\"?|[^"\\])*"|'(?:\\'?|[^'\\])*'|\((?:\\\)?|[^)\\])*\)/).getRegex();
  var reflink = edit(/^!?\[(label)\]\[(ref)\]/).replace("label", _inlineLabel).replace("ref", _blockLabel).getRegex();
  var nolink = edit(/^!?\[(ref)\](?:\[\])?/).replace("ref", _blockLabel).getRegex();
  var reflinkSearch = edit("reflink|nolink(?!\\()", "g").replace("reflink", reflink).replace("nolink", nolink).getRegex();
  var inlineNormal = {
    _backpedal: noopTest,
    // only used for GFM url
    anyPunctuation,
    autolink,
    blockSkip,
    br,
    code: inlineCode,
    del: noopTest,
    emStrongLDelim,
    emStrongRDelimAst,
    emStrongRDelimUnd,
    escape,
    link,
    nolink,
    punctuation,
    reflink,
    reflinkSearch,
    tag,
    text: inlineText,
    url: noopTest
  };
  var inlinePedantic = {
    ...inlineNormal,
    link: edit(/^!?\[(label)\]\((.*?)\)/).replace("label", _inlineLabel).getRegex(),
    reflink: edit(/^!?\[(label)\]\s*\[([^\]]*)\]/).replace("label", _inlineLabel).getRegex()
  };
  var inlineGfm = {
    ...inlineNormal,
    emStrongRDelimAst: emStrongRDelimAstGfm,
    emStrongLDelim: emStrongLDelimGfm,
    url: edit(/^((?:ftp|https?):\/\/|www\.)(?:[a-zA-Z0-9\-]+\.?)+[^\s<]*|^email/, "i").replace("email", /[A-Za-z0-9._+-]+(@)[a-zA-Z0-9-_]+(?:\.[a-zA-Z0-9-_]*[a-zA-Z0-9])+(?![-_])/).getRegex(),
    _backpedal: /(?:[^?!.,:;*_'"~()&]+|\([^)]*\)|&(?![a-zA-Z0-9]+;$)|[?!.,:;*_'"~)]+(?!$))+/,
    del: /^(~~?)(?=[^\s~])((?:\\.|[^\\])*?(?:\\.|[^\s~\\]))\1(?=[^~]|$)/,
    text: /^([`~]+|[^`~])(?:(?= {2,}\n)|(?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)|[\s\S]*?(?:(?=[\\<!\[`*~_]|\b_|https?:\/\/|ftp:\/\/|www\.|$)|[^ ](?= {2,}\n)|[^a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-](?=[a-zA-Z0-9.!#$%&'*+\/=?_`{\|}~-]+@)))/
  };
  var inlineBreaks = {
    ...inlineGfm,
    br: edit(br).replace("{2,}", "*").getRegex(),
    text: edit(inlineGfm.text).replace("\\b_", "\\b_| {2,}\\n").replace(/\{2,\}/g, "*").getRegex()
  };
  var block = {
    normal: blockNormal,
    gfm: blockGfm,
    pedantic: blockPedantic
  };
  var inline = {
    normal: inlineNormal,
    gfm: inlineGfm,
    breaks: inlineBreaks,
    pedantic: inlinePedantic
  };
  var escapeReplacements = {
    "&": "&amp;",
    "<": "&lt;",
    ">": "&gt;",
    '"': "&quot;",
    "'": "&#39;"
  };
  var getEscapeReplacement = (ch) => escapeReplacements[ch];
  function escape2(html22, encode) {
    if (encode) {
      if (other.escapeTest.test(html22)) {
        return html22.replace(other.escapeReplace, getEscapeReplacement);
      }
    } else {
      if (other.escapeTestNoEncode.test(html22)) {
        return html22.replace(other.escapeReplaceNoEncode, getEscapeReplacement);
      }
    }
    return html22;
  }
  function cleanUrl(href) {
    try {
      href = encodeURI(href).replace(other.percentDecode, "%");
    } catch {
      return null;
    }
    return href;
  }
  function splitCells(tableRow, count) {
    const row = tableRow.replace(other.findPipe, (match, offset, str2) => {
      let escaped = false;
      let curr = offset;
      while (--curr >= 0 && str2[curr] === "\\") escaped = !escaped;
      if (escaped) {
        return "|";
      } else {
        return " |";
      }
    }), cells = row.split(other.splitPipe);
    let i = 0;
    if (!cells[0].trim()) {
      cells.shift();
    }
    if (cells.length > 0 && !cells.at(-1)?.trim()) {
      cells.pop();
    }
    if (count) {
      if (cells.length > count) {
        cells.splice(count);
      } else {
        while (cells.length < count) cells.push("");
      }
    }
    for (; i < cells.length; i++) {
      cells[i] = cells[i].trim().replace(other.slashPipe, "|");
    }
    return cells;
  }
  function rtrim(str2, c, invert) {
    const l = str2.length;
    if (l === 0) {
      return "";
    }
    let suffLen = 0;
    while (suffLen < l) {
      const currChar = str2.charAt(l - suffLen - 1);
      if (currChar === c && !invert) {
        suffLen++;
      } else if (currChar !== c && invert) {
        suffLen++;
      } else {
        break;
      }
    }
    return str2.slice(0, l - suffLen);
  }
  function findClosingBracket(str2, b) {
    if (str2.indexOf(b[1]) === -1) {
      return -1;
    }
    let level = 0;
    for (let i = 0; i < str2.length; i++) {
      if (str2[i] === "\\") {
        i++;
      } else if (str2[i] === b[0]) {
        level++;
      } else if (str2[i] === b[1]) {
        level--;
        if (level < 0) {
          return i;
        }
      }
    }
    if (level > 0) {
      return -2;
    }
    return -1;
  }
  function outputLink(cap, link2, raw, lexer2, rules) {
    const href = link2.href;
    const title = link2.title || null;
    const text2 = cap[1].replace(rules.other.outputLinkReplace, "$1");
    lexer2.state.inLink = true;
    const token = {
      type: cap[0].charAt(0) === "!" ? "image" : "link",
      raw,
      href,
      title,
      text: text2,
      tokens: lexer2.inlineTokens(text2)
    };
    lexer2.state.inLink = false;
    return token;
  }
  function indentCodeCompensation(raw, text2, rules) {
    const matchIndentToCode = raw.match(rules.other.indentCodeCompensation);
    if (matchIndentToCode === null) {
      return text2;
    }
    const indentToCode = matchIndentToCode[1];
    return text2.split("\n").map((node) => {
      const matchIndentInNode = node.match(rules.other.beginningSpace);
      if (matchIndentInNode === null) {
        return node;
      }
      const [indentInNode] = matchIndentInNode;
      if (indentInNode.length >= indentToCode.length) {
        return node.slice(indentToCode.length);
      }
      return node;
    }).join("\n");
  }
  var _Tokenizer = class {
    // set by the lexer
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "rules");
      // set by the lexer
      __publicField(this, "lexer");
      this.options = options2 || _defaults;
    }
    space(src) {
      const cap = this.rules.block.newline.exec(src);
      if (cap && cap[0].length > 0) {
        return {
          type: "space",
          raw: cap[0]
        };
      }
    }
    code(src) {
      const cap = this.rules.block.code.exec(src);
      if (cap) {
        const text2 = cap[0].replace(this.rules.other.codeRemoveIndent, "");
        return {
          type: "code",
          raw: cap[0],
          codeBlockStyle: "indented",
          text: !this.options.pedantic ? rtrim(text2, "\n") : text2
        };
      }
    }
    fences(src) {
      const cap = this.rules.block.fences.exec(src);
      if (cap) {
        const raw = cap[0];
        const text2 = indentCodeCompensation(raw, cap[3] || "", this.rules);
        return {
          type: "code",
          raw,
          lang: cap[2] ? cap[2].trim().replace(this.rules.inline.anyPunctuation, "$1") : cap[2],
          text: text2
        };
      }
    }
    heading(src) {
      const cap = this.rules.block.heading.exec(src);
      if (cap) {
        let text2 = cap[2].trim();
        if (this.rules.other.endingHash.test(text2)) {
          const trimmed = rtrim(text2, "#");
          if (this.options.pedantic) {
            text2 = trimmed.trim();
          } else if (!trimmed || this.rules.other.endingSpaceChar.test(trimmed)) {
            text2 = trimmed.trim();
          }
        }
        return {
          type: "heading",
          raw: cap[0],
          depth: cap[1].length,
          text: text2,
          tokens: this.lexer.inline(text2)
        };
      }
    }
    hr(src) {
      const cap = this.rules.block.hr.exec(src);
      if (cap) {
        return {
          type: "hr",
          raw: rtrim(cap[0], "\n")
        };
      }
    }
    blockquote(src) {
      const cap = this.rules.block.blockquote.exec(src);
      if (cap) {
        let lines = rtrim(cap[0], "\n").split("\n");
        let raw = "";
        let text2 = "";
        const tokens = [];
        while (lines.length > 0) {
          let inBlockquote = false;
          const currentLines = [];
          let i;
          for (i = 0; i < lines.length; i++) {
            if (this.rules.other.blockquoteStart.test(lines[i])) {
              currentLines.push(lines[i]);
              inBlockquote = true;
            } else if (!inBlockquote) {
              currentLines.push(lines[i]);
            } else {
              break;
            }
          }
          lines = lines.slice(i);
          const currentRaw = currentLines.join("\n");
          const currentText = currentRaw.replace(this.rules.other.blockquoteSetextReplace, "\n    $1").replace(this.rules.other.blockquoteSetextReplace2, "");
          raw = raw ? `${raw}
${currentRaw}` : currentRaw;
          text2 = text2 ? `${text2}
${currentText}` : currentText;
          const top = this.lexer.state.top;
          this.lexer.state.top = true;
          this.lexer.blockTokens(currentText, tokens, true);
          this.lexer.state.top = top;
          if (lines.length === 0) {
            break;
          }
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "code") {
            break;
          } else if (lastToken?.type === "blockquote") {
            const oldToken = lastToken;
            const newText = oldToken.raw + "\n" + lines.join("\n");
            const newToken = this.blockquote(newText);
            tokens[tokens.length - 1] = newToken;
            raw = raw.substring(0, raw.length - oldToken.raw.length) + newToken.raw;
            text2 = text2.substring(0, text2.length - oldToken.text.length) + newToken.text;
            break;
          } else if (lastToken?.type === "list") {
            const oldToken = lastToken;
            const newText = oldToken.raw + "\n" + lines.join("\n");
            const newToken = this.list(newText);
            tokens[tokens.length - 1] = newToken;
            raw = raw.substring(0, raw.length - lastToken.raw.length) + newToken.raw;
            text2 = text2.substring(0, text2.length - oldToken.raw.length) + newToken.raw;
            lines = newText.substring(tokens.at(-1).raw.length).split("\n");
            continue;
          }
        }
        return {
          type: "blockquote",
          raw,
          tokens,
          text: text2
        };
      }
    }
    list(src) {
      let cap = this.rules.block.list.exec(src);
      if (cap) {
        let bull = cap[1].trim();
        const isordered = bull.length > 1;
        const list2 = {
          type: "list",
          raw: "",
          ordered: isordered,
          start: isordered ? +bull.slice(0, -1) : "",
          loose: false,
          items: []
        };
        bull = isordered ? `\\d{1,9}\\${bull.slice(-1)}` : `\\${bull}`;
        if (this.options.pedantic) {
          bull = isordered ? bull : "[*+-]";
        }
        const itemRegex = this.rules.other.listItemRegex(bull);
        let endsWithBlankLine = false;
        while (src) {
          let endEarly = false;
          let raw = "";
          let itemContents = "";
          if (!(cap = itemRegex.exec(src))) {
            break;
          }
          if (this.rules.block.hr.test(src)) {
            break;
          }
          raw = cap[0];
          src = src.substring(raw.length);
          let line = cap[2].split("\n", 1)[0].replace(this.rules.other.listReplaceTabs, (t2) => " ".repeat(3 * t2.length));
          let nextLine = src.split("\n", 1)[0];
          let blankLine = !line.trim();
          let indent = 0;
          if (this.options.pedantic) {
            indent = 2;
            itemContents = line.trimStart();
          } else if (blankLine) {
            indent = cap[1].length + 1;
          } else {
            indent = cap[2].search(this.rules.other.nonSpaceChar);
            indent = indent > 4 ? 1 : indent;
            itemContents = line.slice(indent);
            indent += cap[1].length;
          }
          if (blankLine && this.rules.other.blankLine.test(nextLine)) {
            raw += nextLine + "\n";
            src = src.substring(nextLine.length + 1);
            endEarly = true;
          }
          if (!endEarly) {
            const nextBulletRegex = this.rules.other.nextBulletRegex(indent);
            const hrRegex = this.rules.other.hrRegex(indent);
            const fencesBeginRegex = this.rules.other.fencesBeginRegex(indent);
            const headingBeginRegex = this.rules.other.headingBeginRegex(indent);
            const htmlBeginRegex = this.rules.other.htmlBeginRegex(indent);
            while (src) {
              const rawLine = src.split("\n", 1)[0];
              let nextLineWithoutTabs;
              nextLine = rawLine;
              if (this.options.pedantic) {
                nextLine = nextLine.replace(this.rules.other.listReplaceNesting, "  ");
                nextLineWithoutTabs = nextLine;
              } else {
                nextLineWithoutTabs = nextLine.replace(this.rules.other.tabCharGlobal, "    ");
              }
              if (fencesBeginRegex.test(nextLine)) {
                break;
              }
              if (headingBeginRegex.test(nextLine)) {
                break;
              }
              if (htmlBeginRegex.test(nextLine)) {
                break;
              }
              if (nextBulletRegex.test(nextLine)) {
                break;
              }
              if (hrRegex.test(nextLine)) {
                break;
              }
              if (nextLineWithoutTabs.search(this.rules.other.nonSpaceChar) >= indent || !nextLine.trim()) {
                itemContents += "\n" + nextLineWithoutTabs.slice(indent);
              } else {
                if (blankLine) {
                  break;
                }
                if (line.replace(this.rules.other.tabCharGlobal, "    ").search(this.rules.other.nonSpaceChar) >= 4) {
                  break;
                }
                if (fencesBeginRegex.test(line)) {
                  break;
                }
                if (headingBeginRegex.test(line)) {
                  break;
                }
                if (hrRegex.test(line)) {
                  break;
                }
                itemContents += "\n" + nextLine;
              }
              if (!blankLine && !nextLine.trim()) {
                blankLine = true;
              }
              raw += rawLine + "\n";
              src = src.substring(rawLine.length + 1);
              line = nextLineWithoutTabs.slice(indent);
            }
          }
          if (!list2.loose) {
            if (endsWithBlankLine) {
              list2.loose = true;
            } else if (this.rules.other.doubleBlankLine.test(raw)) {
              endsWithBlankLine = true;
            }
          }
          let istask = null;
          let ischecked;
          if (this.options.gfm) {
            istask = this.rules.other.listIsTask.exec(itemContents);
            if (istask) {
              ischecked = istask[0] !== "[ ] ";
              itemContents = itemContents.replace(this.rules.other.listReplaceTask, "");
            }
          }
          list2.items.push({
            type: "list_item",
            raw,
            task: !!istask,
            checked: ischecked,
            loose: false,
            text: itemContents,
            tokens: []
          });
          list2.raw += raw;
        }
        const lastItem = list2.items.at(-1);
        if (lastItem) {
          lastItem.raw = lastItem.raw.trimEnd();
          lastItem.text = lastItem.text.trimEnd();
        } else {
          return;
        }
        list2.raw = list2.raw.trimEnd();
        for (let i = 0; i < list2.items.length; i++) {
          this.lexer.state.top = false;
          list2.items[i].tokens = this.lexer.blockTokens(list2.items[i].text, []);
          if (!list2.loose) {
            const spacers = list2.items[i].tokens.filter((t2) => t2.type === "space");
            const hasMultipleLineBreaks = spacers.length > 0 && spacers.some((t2) => this.rules.other.anyLine.test(t2.raw));
            list2.loose = hasMultipleLineBreaks;
          }
        }
        if (list2.loose) {
          for (let i = 0; i < list2.items.length; i++) {
            list2.items[i].loose = true;
          }
        }
        return list2;
      }
    }
    html(src) {
      const cap = this.rules.block.html.exec(src);
      if (cap) {
        const token = {
          type: "html",
          block: true,
          raw: cap[0],
          pre: cap[1] === "pre" || cap[1] === "script" || cap[1] === "style",
          text: cap[0]
        };
        return token;
      }
    }
    def(src) {
      const cap = this.rules.block.def.exec(src);
      if (cap) {
        const tag2 = cap[1].toLowerCase().replace(this.rules.other.multipleSpaceGlobal, " ");
        const href = cap[2] ? cap[2].replace(this.rules.other.hrefBrackets, "$1").replace(this.rules.inline.anyPunctuation, "$1") : "";
        const title = cap[3] ? cap[3].substring(1, cap[3].length - 1).replace(this.rules.inline.anyPunctuation, "$1") : cap[3];
        return {
          type: "def",
          tag: tag2,
          raw: cap[0],
          href,
          title
        };
      }
    }
    table(src) {
      const cap = this.rules.block.table.exec(src);
      if (!cap) {
        return;
      }
      if (!this.rules.other.tableDelimiter.test(cap[2])) {
        return;
      }
      const headers = splitCells(cap[1]);
      const aligns = cap[2].replace(this.rules.other.tableAlignChars, "").split("|");
      const rows = cap[3]?.trim() ? cap[3].replace(this.rules.other.tableRowBlankLine, "").split("\n") : [];
      const item = {
        type: "table",
        raw: cap[0],
        header: [],
        align: [],
        rows: []
      };
      if (headers.length !== aligns.length) {
        return;
      }
      for (const align of aligns) {
        if (this.rules.other.tableAlignRight.test(align)) {
          item.align.push("right");
        } else if (this.rules.other.tableAlignCenter.test(align)) {
          item.align.push("center");
        } else if (this.rules.other.tableAlignLeft.test(align)) {
          item.align.push("left");
        } else {
          item.align.push(null);
        }
      }
      for (let i = 0; i < headers.length; i++) {
        item.header.push({
          text: headers[i],
          tokens: this.lexer.inline(headers[i]),
          header: true,
          align: item.align[i]
        });
      }
      for (const row of rows) {
        item.rows.push(splitCells(row, item.header.length).map((cell, i) => {
          return {
            text: cell,
            tokens: this.lexer.inline(cell),
            header: false,
            align: item.align[i]
          };
        }));
      }
      return item;
    }
    lheading(src) {
      const cap = this.rules.block.lheading.exec(src);
      if (cap) {
        return {
          type: "heading",
          raw: cap[0],
          depth: cap[2].charAt(0) === "=" ? 1 : 2,
          text: cap[1],
          tokens: this.lexer.inline(cap[1])
        };
      }
    }
    paragraph(src) {
      const cap = this.rules.block.paragraph.exec(src);
      if (cap) {
        const text2 = cap[1].charAt(cap[1].length - 1) === "\n" ? cap[1].slice(0, -1) : cap[1];
        return {
          type: "paragraph",
          raw: cap[0],
          text: text2,
          tokens: this.lexer.inline(text2)
        };
      }
    }
    text(src) {
      const cap = this.rules.block.text.exec(src);
      if (cap) {
        return {
          type: "text",
          raw: cap[0],
          text: cap[0],
          tokens: this.lexer.inline(cap[0])
        };
      }
    }
    escape(src) {
      const cap = this.rules.inline.escape.exec(src);
      if (cap) {
        return {
          type: "escape",
          raw: cap[0],
          text: cap[1]
        };
      }
    }
    tag(src) {
      const cap = this.rules.inline.tag.exec(src);
      if (cap) {
        if (!this.lexer.state.inLink && this.rules.other.startATag.test(cap[0])) {
          this.lexer.state.inLink = true;
        } else if (this.lexer.state.inLink && this.rules.other.endATag.test(cap[0])) {
          this.lexer.state.inLink = false;
        }
        if (!this.lexer.state.inRawBlock && this.rules.other.startPreScriptTag.test(cap[0])) {
          this.lexer.state.inRawBlock = true;
        } else if (this.lexer.state.inRawBlock && this.rules.other.endPreScriptTag.test(cap[0])) {
          this.lexer.state.inRawBlock = false;
        }
        return {
          type: "html",
          raw: cap[0],
          inLink: this.lexer.state.inLink,
          inRawBlock: this.lexer.state.inRawBlock,
          block: false,
          text: cap[0]
        };
      }
    }
    link(src) {
      const cap = this.rules.inline.link.exec(src);
      if (cap) {
        const trimmedUrl = cap[2].trim();
        if (!this.options.pedantic && this.rules.other.startAngleBracket.test(trimmedUrl)) {
          if (!this.rules.other.endAngleBracket.test(trimmedUrl)) {
            return;
          }
          const rtrimSlash = rtrim(trimmedUrl.slice(0, -1), "\\");
          if ((trimmedUrl.length - rtrimSlash.length) % 2 === 0) {
            return;
          }
        } else {
          const lastParenIndex = findClosingBracket(cap[2], "()");
          if (lastParenIndex === -2) {
            return;
          }
          if (lastParenIndex > -1) {
            const start = cap[0].indexOf("!") === 0 ? 5 : 4;
            const linkLen = start + cap[1].length + lastParenIndex;
            cap[2] = cap[2].substring(0, lastParenIndex);
            cap[0] = cap[0].substring(0, linkLen).trim();
            cap[3] = "";
          }
        }
        let href = cap[2];
        let title = "";
        if (this.options.pedantic) {
          const link2 = this.rules.other.pedanticHrefTitle.exec(href);
          if (link2) {
            href = link2[1];
            title = link2[3];
          }
        } else {
          title = cap[3] ? cap[3].slice(1, -1) : "";
        }
        href = href.trim();
        if (this.rules.other.startAngleBracket.test(href)) {
          if (this.options.pedantic && !this.rules.other.endAngleBracket.test(trimmedUrl)) {
            href = href.slice(1);
          } else {
            href = href.slice(1, -1);
          }
        }
        return outputLink(cap, {
          href: href ? href.replace(this.rules.inline.anyPunctuation, "$1") : href,
          title: title ? title.replace(this.rules.inline.anyPunctuation, "$1") : title
        }, cap[0], this.lexer, this.rules);
      }
    }
    reflink(src, links) {
      let cap;
      if ((cap = this.rules.inline.reflink.exec(src)) || (cap = this.rules.inline.nolink.exec(src))) {
        const linkString = (cap[2] || cap[1]).replace(this.rules.other.multipleSpaceGlobal, " ");
        const link2 = links[linkString.toLowerCase()];
        if (!link2) {
          const text2 = cap[0].charAt(0);
          return {
            type: "text",
            raw: text2,
            text: text2
          };
        }
        return outputLink(cap, link2, cap[0], this.lexer, this.rules);
      }
    }
    emStrong(src, maskedSrc, prevChar = "") {
      let match = this.rules.inline.emStrongLDelim.exec(src);
      if (!match) return;
      if (match[3] && prevChar.match(this.rules.other.unicodeAlphaNumeric)) return;
      const nextChar = match[1] || match[2] || "";
      if (!nextChar || !prevChar || this.rules.inline.punctuation.exec(prevChar)) {
        const lLength = [...match[0]].length - 1;
        let rDelim, rLength, delimTotal = lLength, midDelimTotal = 0;
        const endReg = match[0][0] === "*" ? this.rules.inline.emStrongRDelimAst : this.rules.inline.emStrongRDelimUnd;
        endReg.lastIndex = 0;
        maskedSrc = maskedSrc.slice(-1 * src.length + lLength);
        while ((match = endReg.exec(maskedSrc)) != null) {
          rDelim = match[1] || match[2] || match[3] || match[4] || match[5] || match[6];
          if (!rDelim) continue;
          rLength = [...rDelim].length;
          if (match[3] || match[4]) {
            delimTotal += rLength;
            continue;
          } else if (match[5] || match[6]) {
            if (lLength % 3 && !((lLength + rLength) % 3)) {
              midDelimTotal += rLength;
              continue;
            }
          }
          delimTotal -= rLength;
          if (delimTotal > 0) continue;
          rLength = Math.min(rLength, rLength + delimTotal + midDelimTotal);
          const lastCharLength = [...match[0]][0].length;
          const raw = src.slice(0, lLength + match.index + lastCharLength + rLength);
          if (Math.min(lLength, rLength) % 2) {
            const text22 = raw.slice(1, -1);
            return {
              type: "em",
              raw,
              text: text22,
              tokens: this.lexer.inlineTokens(text22)
            };
          }
          const text2 = raw.slice(2, -2);
          return {
            type: "strong",
            raw,
            text: text2,
            tokens: this.lexer.inlineTokens(text2)
          };
        }
      }
    }
    codespan(src) {
      const cap = this.rules.inline.code.exec(src);
      if (cap) {
        let text2 = cap[2].replace(this.rules.other.newLineCharGlobal, " ");
        const hasNonSpaceChars = this.rules.other.nonSpaceChar.test(text2);
        const hasSpaceCharsOnBothEnds = this.rules.other.startingSpaceChar.test(text2) && this.rules.other.endingSpaceChar.test(text2);
        if (hasNonSpaceChars && hasSpaceCharsOnBothEnds) {
          text2 = text2.substring(1, text2.length - 1);
        }
        return {
          type: "codespan",
          raw: cap[0],
          text: text2
        };
      }
    }
    br(src) {
      const cap = this.rules.inline.br.exec(src);
      if (cap) {
        return {
          type: "br",
          raw: cap[0]
        };
      }
    }
    del(src) {
      const cap = this.rules.inline.del.exec(src);
      if (cap) {
        return {
          type: "del",
          raw: cap[0],
          text: cap[2],
          tokens: this.lexer.inlineTokens(cap[2])
        };
      }
    }
    autolink(src) {
      const cap = this.rules.inline.autolink.exec(src);
      if (cap) {
        let text2, href;
        if (cap[2] === "@") {
          text2 = cap[1];
          href = "mailto:" + text2;
        } else {
          text2 = cap[1];
          href = text2;
        }
        return {
          type: "link",
          raw: cap[0],
          text: text2,
          href,
          tokens: [
            {
              type: "text",
              raw: text2,
              text: text2
            }
          ]
        };
      }
    }
    url(src) {
      let cap;
      if (cap = this.rules.inline.url.exec(src)) {
        let text2, href;
        if (cap[2] === "@") {
          text2 = cap[0];
          href = "mailto:" + text2;
        } else {
          let prevCapZero;
          do {
            prevCapZero = cap[0];
            cap[0] = this.rules.inline._backpedal.exec(cap[0])?.[0] ?? "";
          } while (prevCapZero !== cap[0]);
          text2 = cap[0];
          if (cap[1] === "www.") {
            href = "http://" + cap[0];
          } else {
            href = cap[0];
          }
        }
        return {
          type: "link",
          raw: cap[0],
          text: text2,
          href,
          tokens: [
            {
              type: "text",
              raw: text2,
              text: text2
            }
          ]
        };
      }
    }
    inlineText(src) {
      const cap = this.rules.inline.text.exec(src);
      if (cap) {
        const escaped = this.lexer.state.inRawBlock;
        return {
          type: "text",
          raw: cap[0],
          text: cap[0],
          escaped
        };
      }
    }
  };
  var _Lexer = class __Lexer {
    constructor(options2) {
      __publicField(this, "tokens");
      __publicField(this, "options");
      __publicField(this, "state");
      __publicField(this, "tokenizer");
      __publicField(this, "inlineQueue");
      this.tokens = [];
      this.tokens.links = /* @__PURE__ */ Object.create(null);
      this.options = options2 || _defaults;
      this.options.tokenizer = this.options.tokenizer || new _Tokenizer();
      this.tokenizer = this.options.tokenizer;
      this.tokenizer.options = this.options;
      this.tokenizer.lexer = this;
      this.inlineQueue = [];
      this.state = {
        inLink: false,
        inRawBlock: false,
        top: true
      };
      const rules = {
        other,
        block: block.normal,
        inline: inline.normal
      };
      if (this.options.pedantic) {
        rules.block = block.pedantic;
        rules.inline = inline.pedantic;
      } else if (this.options.gfm) {
        rules.block = block.gfm;
        if (this.options.breaks) {
          rules.inline = inline.breaks;
        } else {
          rules.inline = inline.gfm;
        }
      }
      this.tokenizer.rules = rules;
    }
    /**
     * Expose Rules
     */
    static get rules() {
      return {
        block,
        inline
      };
    }
    /**
     * Static Lex Method
     */
    static lex(src, options2) {
      const lexer2 = new __Lexer(options2);
      return lexer2.lex(src);
    }
    /**
     * Static Lex Inline Method
     */
    static lexInline(src, options2) {
      const lexer2 = new __Lexer(options2);
      return lexer2.inlineTokens(src);
    }
    /**
     * Preprocessing
     */
    lex(src) {
      src = src.replace(other.carriageReturn, "\n");
      this.blockTokens(src, this.tokens);
      for (let i = 0; i < this.inlineQueue.length; i++) {
        const next = this.inlineQueue[i];
        this.inlineTokens(next.src, next.tokens);
      }
      this.inlineQueue = [];
      return this.tokens;
    }
    blockTokens(src, tokens = [], lastParagraphClipped = false) {
      if (this.options.pedantic) {
        src = src.replace(other.tabCharGlobal, "    ").replace(other.spaceLine, "");
      }
      while (src) {
        let token;
        if (this.options.extensions?.block?.some((extTokenizer) => {
          if (token = extTokenizer.call({ lexer: this }, src, tokens)) {
            src = src.substring(token.raw.length);
            tokens.push(token);
            return true;
          }
          return false;
        })) {
          continue;
        }
        if (token = this.tokenizer.space(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (token.raw.length === 1 && lastToken !== void 0) {
            lastToken.raw += "\n";
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (token = this.tokenizer.code(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "paragraph" || lastToken?.type === "text") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.text;
            this.inlineQueue.at(-1).src = lastToken.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (token = this.tokenizer.fences(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.heading(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.hr(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.blockquote(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.list(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.html(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.def(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "paragraph" || lastToken?.type === "text") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.raw;
            this.inlineQueue.at(-1).src = lastToken.text;
          } else if (!this.tokens.links[token.tag]) {
            this.tokens.links[token.tag] = {
              href: token.href,
              title: token.title
            };
          }
          continue;
        }
        if (token = this.tokenizer.table(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.lheading(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        let cutSrc = src;
        if (this.options.extensions?.startBlock) {
          let startIndex = Infinity;
          const tempSrc = src.slice(1);
          let tempStart;
          this.options.extensions.startBlock.forEach((getStartIndex) => {
            tempStart = getStartIndex.call({ lexer: this }, tempSrc);
            if (typeof tempStart === "number" && tempStart >= 0) {
              startIndex = Math.min(startIndex, tempStart);
            }
          });
          if (startIndex < Infinity && startIndex >= 0) {
            cutSrc = src.substring(0, startIndex + 1);
          }
        }
        if (this.state.top && (token = this.tokenizer.paragraph(cutSrc))) {
          const lastToken = tokens.at(-1);
          if (lastParagraphClipped && lastToken?.type === "paragraph") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.text;
            this.inlineQueue.pop();
            this.inlineQueue.at(-1).src = lastToken.text;
          } else {
            tokens.push(token);
          }
          lastParagraphClipped = cutSrc.length !== src.length;
          src = src.substring(token.raw.length);
          continue;
        }
        if (token = this.tokenizer.text(src)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "text") {
            lastToken.raw += "\n" + token.raw;
            lastToken.text += "\n" + token.text;
            this.inlineQueue.pop();
            this.inlineQueue.at(-1).src = lastToken.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (src) {
          const errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
          if (this.options.silent) {
            console.error(errMsg);
            break;
          } else {
            throw new Error(errMsg);
          }
        }
      }
      this.state.top = true;
      return tokens;
    }
    inline(src, tokens = []) {
      this.inlineQueue.push({ src, tokens });
      return tokens;
    }
    /**
     * Lexing/Compiling
     */
    inlineTokens(src, tokens = []) {
      let maskedSrc = src;
      let match = null;
      if (this.tokens.links) {
        const links = Object.keys(this.tokens.links);
        if (links.length > 0) {
          while ((match = this.tokenizer.rules.inline.reflinkSearch.exec(maskedSrc)) != null) {
            if (links.includes(match[0].slice(match[0].lastIndexOf("[") + 1, -1))) {
              maskedSrc = maskedSrc.slice(0, match.index) + "[" + "a".repeat(match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.reflinkSearch.lastIndex);
            }
          }
        }
      }
      while ((match = this.tokenizer.rules.inline.anyPunctuation.exec(maskedSrc)) != null) {
        maskedSrc = maskedSrc.slice(0, match.index) + "++" + maskedSrc.slice(this.tokenizer.rules.inline.anyPunctuation.lastIndex);
      }
      while ((match = this.tokenizer.rules.inline.blockSkip.exec(maskedSrc)) != null) {
        maskedSrc = maskedSrc.slice(0, match.index) + "[" + "a".repeat(match[0].length - 2) + "]" + maskedSrc.slice(this.tokenizer.rules.inline.blockSkip.lastIndex);
      }
      let keepPrevChar = false;
      let prevChar = "";
      while (src) {
        if (!keepPrevChar) {
          prevChar = "";
        }
        keepPrevChar = false;
        let token;
        if (this.options.extensions?.inline?.some((extTokenizer) => {
          if (token = extTokenizer.call({ lexer: this }, src, tokens)) {
            src = src.substring(token.raw.length);
            tokens.push(token);
            return true;
          }
          return false;
        })) {
          continue;
        }
        if (token = this.tokenizer.escape(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.tag(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.link(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.reflink(src, this.tokens.links)) {
          src = src.substring(token.raw.length);
          const lastToken = tokens.at(-1);
          if (token.type === "text" && lastToken?.type === "text") {
            lastToken.raw += token.raw;
            lastToken.text += token.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (token = this.tokenizer.emStrong(src, maskedSrc, prevChar)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.codespan(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.br(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.del(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (token = this.tokenizer.autolink(src)) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        if (!this.state.inLink && (token = this.tokenizer.url(src))) {
          src = src.substring(token.raw.length);
          tokens.push(token);
          continue;
        }
        let cutSrc = src;
        if (this.options.extensions?.startInline) {
          let startIndex = Infinity;
          const tempSrc = src.slice(1);
          let tempStart;
          this.options.extensions.startInline.forEach((getStartIndex) => {
            tempStart = getStartIndex.call({ lexer: this }, tempSrc);
            if (typeof tempStart === "number" && tempStart >= 0) {
              startIndex = Math.min(startIndex, tempStart);
            }
          });
          if (startIndex < Infinity && startIndex >= 0) {
            cutSrc = src.substring(0, startIndex + 1);
          }
        }
        if (token = this.tokenizer.inlineText(cutSrc)) {
          src = src.substring(token.raw.length);
          if (token.raw.slice(-1) !== "_") {
            prevChar = token.raw.slice(-1);
          }
          keepPrevChar = true;
          const lastToken = tokens.at(-1);
          if (lastToken?.type === "text") {
            lastToken.raw += token.raw;
            lastToken.text += token.text;
          } else {
            tokens.push(token);
          }
          continue;
        }
        if (src) {
          const errMsg = "Infinite loop on byte: " + src.charCodeAt(0);
          if (this.options.silent) {
            console.error(errMsg);
            break;
          } else {
            throw new Error(errMsg);
          }
        }
      }
      return tokens;
    }
  };
  var _Renderer = class {
    // set by the parser
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "parser");
      this.options = options2 || _defaults;
    }
    space(token) {
      return "";
    }
    code({ text: text2, lang, escaped }) {
      const langString = (lang || "").match(other.notSpaceStart)?.[0];
      const code = text2.replace(other.endingNewline, "") + "\n";
      if (!langString) {
        return "<pre><code>" + (escaped ? code : escape2(code, true)) + "</code></pre>\n";
      }
      return '<pre><code class="language-' + escape2(langString) + '">' + (escaped ? code : escape2(code, true)) + "</code></pre>\n";
    }
    blockquote({ tokens }) {
      const body = this.parser.parse(tokens);
      return `<blockquote>
${body}</blockquote>
`;
    }
    html({ text: text2 }) {
      return text2;
    }
    heading({ tokens, depth }) {
      return `<h${depth}>${this.parser.parseInline(tokens)}</h${depth}>
`;
    }
    hr(token) {
      return "<hr>\n";
    }
    list(token) {
      const ordered = token.ordered;
      const start = token.start;
      let body = "";
      for (let j = 0; j < token.items.length; j++) {
        const item = token.items[j];
        body += this.listitem(item);
      }
      const type2 = ordered ? "ol" : "ul";
      const startAttr = ordered && start !== 1 ? ' start="' + start + '"' : "";
      return "<" + type2 + startAttr + ">\n" + body + "</" + type2 + ">\n";
    }
    listitem(item) {
      let itemBody = "";
      if (item.task) {
        const checkbox = this.checkbox({ checked: !!item.checked });
        if (item.loose) {
          if (item.tokens[0]?.type === "paragraph") {
            item.tokens[0].text = checkbox + " " + item.tokens[0].text;
            if (item.tokens[0].tokens && item.tokens[0].tokens.length > 0 && item.tokens[0].tokens[0].type === "text") {
              item.tokens[0].tokens[0].text = checkbox + " " + escape2(item.tokens[0].tokens[0].text);
              item.tokens[0].tokens[0].escaped = true;
            }
          } else {
            item.tokens.unshift({
              type: "text",
              raw: checkbox + " ",
              text: checkbox + " ",
              escaped: true
            });
          }
        } else {
          itemBody += checkbox + " ";
        }
      }
      itemBody += this.parser.parse(item.tokens, !!item.loose);
      return `<li>${itemBody}</li>
`;
    }
    checkbox({ checked }) {
      return "<input " + (checked ? 'checked="" ' : "") + 'disabled="" type="checkbox">';
    }
    paragraph({ tokens }) {
      return `<p>${this.parser.parseInline(tokens)}</p>
`;
    }
    table(token) {
      let header = "";
      let cell = "";
      for (let j = 0; j < token.header.length; j++) {
        cell += this.tablecell(token.header[j]);
      }
      header += this.tablerow({ text: cell });
      let body = "";
      for (let j = 0; j < token.rows.length; j++) {
        const row = token.rows[j];
        cell = "";
        for (let k = 0; k < row.length; k++) {
          cell += this.tablecell(row[k]);
        }
        body += this.tablerow({ text: cell });
      }
      if (body) body = `<tbody>${body}</tbody>`;
      return "<table>\n<thead>\n" + header + "</thead>\n" + body + "</table>\n";
    }
    tablerow({ text: text2 }) {
      return `<tr>
${text2}</tr>
`;
    }
    tablecell(token) {
      const content = this.parser.parseInline(token.tokens);
      const type2 = token.header ? "th" : "td";
      const tag2 = token.align ? `<${type2} align="${token.align}">` : `<${type2}>`;
      return tag2 + content + `</${type2}>
`;
    }
    /**
     * span level renderer
     */
    strong({ tokens }) {
      return `<strong>${this.parser.parseInline(tokens)}</strong>`;
    }
    em({ tokens }) {
      return `<em>${this.parser.parseInline(tokens)}</em>`;
    }
    codespan({ text: text2 }) {
      return `<code>${escape2(text2, true)}</code>`;
    }
    br(token) {
      return "<br>";
    }
    del({ tokens }) {
      return `<del>${this.parser.parseInline(tokens)}</del>`;
    }
    link({ href, title, tokens }) {
      const text2 = this.parser.parseInline(tokens);
      const cleanHref = cleanUrl(href);
      if (cleanHref === null) {
        return text2;
      }
      href = cleanHref;
      let out = '<a href="' + href + '"';
      if (title) {
        out += ' title="' + escape2(title) + '"';
      }
      out += ">" + text2 + "</a>";
      return out;
    }
    image({ href, title, text: text2, tokens }) {
      if (tokens) {
        text2 = this.parser.parseInline(tokens, this.parser.textRenderer);
      }
      const cleanHref = cleanUrl(href);
      if (cleanHref === null) {
        return escape2(text2);
      }
      href = cleanHref;
      let out = `<img src="${href}" alt="${text2}"`;
      if (title) {
        out += ` title="${escape2(title)}"`;
      }
      out += ">";
      return out;
    }
    text(token) {
      return "tokens" in token && token.tokens ? this.parser.parseInline(token.tokens) : "escaped" in token && token.escaped ? token.text : escape2(token.text);
    }
  };
  var _TextRenderer = class {
    // no need for block level renderers
    strong({ text: text2 }) {
      return text2;
    }
    em({ text: text2 }) {
      return text2;
    }
    codespan({ text: text2 }) {
      return text2;
    }
    del({ text: text2 }) {
      return text2;
    }
    html({ text: text2 }) {
      return text2;
    }
    text({ text: text2 }) {
      return text2;
    }
    link({ text: text2 }) {
      return "" + text2;
    }
    image({ text: text2 }) {
      return "" + text2;
    }
    br() {
      return "";
    }
  };
  var _Parser = class __Parser {
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "renderer");
      __publicField(this, "textRenderer");
      this.options = options2 || _defaults;
      this.options.renderer = this.options.renderer || new _Renderer();
      this.renderer = this.options.renderer;
      this.renderer.options = this.options;
      this.renderer.parser = this;
      this.textRenderer = new _TextRenderer();
    }
    /**
     * Static Parse Method
     */
    static parse(tokens, options2) {
      const parser2 = new __Parser(options2);
      return parser2.parse(tokens);
    }
    /**
     * Static Parse Inline Method
     */
    static parseInline(tokens, options2) {
      const parser2 = new __Parser(options2);
      return parser2.parseInline(tokens);
    }
    /**
     * Parse Loop
     */
    parse(tokens, top = true) {
      let out = "";
      for (let i = 0; i < tokens.length; i++) {
        const anyToken = tokens[i];
        if (this.options.extensions?.renderers?.[anyToken.type]) {
          const genericToken = anyToken;
          const ret = this.options.extensions.renderers[genericToken.type].call({ parser: this }, genericToken);
          if (ret !== false || !["space", "hr", "heading", "code", "table", "blockquote", "list", "html", "paragraph", "text"].includes(genericToken.type)) {
            out += ret || "";
            continue;
          }
        }
        const token = anyToken;
        switch (token.type) {
          case "space": {
            out += this.renderer.space(token);
            continue;
          }
          case "hr": {
            out += this.renderer.hr(token);
            continue;
          }
          case "heading": {
            out += this.renderer.heading(token);
            continue;
          }
          case "code": {
            out += this.renderer.code(token);
            continue;
          }
          case "table": {
            out += this.renderer.table(token);
            continue;
          }
          case "blockquote": {
            out += this.renderer.blockquote(token);
            continue;
          }
          case "list": {
            out += this.renderer.list(token);
            continue;
          }
          case "html": {
            out += this.renderer.html(token);
            continue;
          }
          case "paragraph": {
            out += this.renderer.paragraph(token);
            continue;
          }
          case "text": {
            let textToken = token;
            let body = this.renderer.text(textToken);
            while (i + 1 < tokens.length && tokens[i + 1].type === "text") {
              textToken = tokens[++i];
              body += "\n" + this.renderer.text(textToken);
            }
            if (top) {
              out += this.renderer.paragraph({
                type: "paragraph",
                raw: body,
                text: body,
                tokens: [{ type: "text", raw: body, text: body, escaped: true }]
              });
            } else {
              out += body;
            }
            continue;
          }
          default: {
            const errMsg = 'Token with "' + token.type + '" type was not found.';
            if (this.options.silent) {
              console.error(errMsg);
              return "";
            } else {
              throw new Error(errMsg);
            }
          }
        }
      }
      return out;
    }
    /**
     * Parse Inline Tokens
     */
    parseInline(tokens, renderer = this.renderer) {
      let out = "";
      for (let i = 0; i < tokens.length; i++) {
        const anyToken = tokens[i];
        if (this.options.extensions?.renderers?.[anyToken.type]) {
          const ret = this.options.extensions.renderers[anyToken.type].call({ parser: this }, anyToken);
          if (ret !== false || !["escape", "html", "link", "image", "strong", "em", "codespan", "br", "del", "text"].includes(anyToken.type)) {
            out += ret || "";
            continue;
          }
        }
        const token = anyToken;
        switch (token.type) {
          case "escape": {
            out += renderer.text(token);
            break;
          }
          case "html": {
            out += renderer.html(token);
            break;
          }
          case "link": {
            out += renderer.link(token);
            break;
          }
          case "image": {
            out += renderer.image(token);
            break;
          }
          case "strong": {
            out += renderer.strong(token);
            break;
          }
          case "em": {
            out += renderer.em(token);
            break;
          }
          case "codespan": {
            out += renderer.codespan(token);
            break;
          }
          case "br": {
            out += renderer.br(token);
            break;
          }
          case "del": {
            out += renderer.del(token);
            break;
          }
          case "text": {
            out += renderer.text(token);
            break;
          }
          default: {
            const errMsg = 'Token with "' + token.type + '" type was not found.';
            if (this.options.silent) {
              console.error(errMsg);
              return "";
            } else {
              throw new Error(errMsg);
            }
          }
        }
      }
      return out;
    }
  };
  var _a;
  var _Hooks = (_a = class {
    constructor(options2) {
      __publicField(this, "options");
      __publicField(this, "block");
      this.options = options2 || _defaults;
    }
    /**
     * Process markdown before marked
     */
    preprocess(markdown) {
      return markdown;
    }
    /**
     * Process HTML after marked is finished
     */
    postprocess(html22) {
      return html22;
    }
    /**
     * Process all tokens before walk tokens
     */
    processAllTokens(tokens) {
      return tokens;
    }
    /**
     * Provide function to tokenize markdown
     */
    provideLexer() {
      return this.block ? _Lexer.lex : _Lexer.lexInline;
    }
    /**
     * Provide function to parse tokens
     */
    provideParser() {
      return this.block ? _Parser.parse : _Parser.parseInline;
    }
  }, __publicField(_a, "passThroughHooks", /* @__PURE__ */ new Set([
    "preprocess",
    "postprocess",
    "processAllTokens"
  ])), _a);
  var Marked = class {
    constructor(...args) {
      __publicField(this, "defaults", _getDefaults());
      __publicField(this, "options", this.setOptions);
      __publicField(this, "parse", this.parseMarkdown(true));
      __publicField(this, "parseInline", this.parseMarkdown(false));
      __publicField(this, "Parser", _Parser);
      __publicField(this, "Renderer", _Renderer);
      __publicField(this, "TextRenderer", _TextRenderer);
      __publicField(this, "Lexer", _Lexer);
      __publicField(this, "Tokenizer", _Tokenizer);
      __publicField(this, "Hooks", _Hooks);
      this.use(...args);
    }
    /**
     * Run callback for every token
     */
    walkTokens(tokens, callback) {
      let values = [];
      for (const token of tokens) {
        values = values.concat(callback.call(this, token));
        switch (token.type) {
          case "table": {
            const tableToken = token;
            for (const cell of tableToken.header) {
              values = values.concat(this.walkTokens(cell.tokens, callback));
            }
            for (const row of tableToken.rows) {
              for (const cell of row) {
                values = values.concat(this.walkTokens(cell.tokens, callback));
              }
            }
            break;
          }
          case "list": {
            const listToken = token;
            values = values.concat(this.walkTokens(listToken.items, callback));
            break;
          }
          default: {
            const genericToken = token;
            if (this.defaults.extensions?.childTokens?.[genericToken.type]) {
              this.defaults.extensions.childTokens[genericToken.type].forEach((childTokens) => {
                const tokens2 = genericToken[childTokens].flat(Infinity);
                values = values.concat(this.walkTokens(tokens2, callback));
              });
            } else if (genericToken.tokens) {
              values = values.concat(this.walkTokens(genericToken.tokens, callback));
            }
          }
        }
      }
      return values;
    }
    use(...args) {
      const extensions = this.defaults.extensions || { renderers: {}, childTokens: {} };
      args.forEach((pack) => {
        const opts = { ...pack };
        opts.async = this.defaults.async || opts.async || false;
        if (pack.extensions) {
          pack.extensions.forEach((ext) => {
            if (!ext.name) {
              throw new Error("extension name required");
            }
            if ("renderer" in ext) {
              const prevRenderer = extensions.renderers[ext.name];
              if (prevRenderer) {
                extensions.renderers[ext.name] = function(...args2) {
                  let ret = ext.renderer.apply(this, args2);
                  if (ret === false) {
                    ret = prevRenderer.apply(this, args2);
                  }
                  return ret;
                };
              } else {
                extensions.renderers[ext.name] = ext.renderer;
              }
            }
            if ("tokenizer" in ext) {
              if (!ext.level || ext.level !== "block" && ext.level !== "inline") {
                throw new Error("extension level must be 'block' or 'inline'");
              }
              const extLevel = extensions[ext.level];
              if (extLevel) {
                extLevel.unshift(ext.tokenizer);
              } else {
                extensions[ext.level] = [ext.tokenizer];
              }
              if (ext.start) {
                if (ext.level === "block") {
                  if (extensions.startBlock) {
                    extensions.startBlock.push(ext.start);
                  } else {
                    extensions.startBlock = [ext.start];
                  }
                } else if (ext.level === "inline") {
                  if (extensions.startInline) {
                    extensions.startInline.push(ext.start);
                  } else {
                    extensions.startInline = [ext.start];
                  }
                }
              }
            }
            if ("childTokens" in ext && ext.childTokens) {
              extensions.childTokens[ext.name] = ext.childTokens;
            }
          });
          opts.extensions = extensions;
        }
        if (pack.renderer) {
          const renderer = this.defaults.renderer || new _Renderer(this.defaults);
          for (const prop in pack.renderer) {
            if (!(prop in renderer)) {
              throw new Error(`renderer '${prop}' does not exist`);
            }
            if (["options", "parser"].includes(prop)) {
              continue;
            }
            const rendererProp = prop;
            const rendererFunc = pack.renderer[rendererProp];
            const prevRenderer = renderer[rendererProp];
            renderer[rendererProp] = (...args2) => {
              let ret = rendererFunc.apply(renderer, args2);
              if (ret === false) {
                ret = prevRenderer.apply(renderer, args2);
              }
              return ret || "";
            };
          }
          opts.renderer = renderer;
        }
        if (pack.tokenizer) {
          const tokenizer = this.defaults.tokenizer || new _Tokenizer(this.defaults);
          for (const prop in pack.tokenizer) {
            if (!(prop in tokenizer)) {
              throw new Error(`tokenizer '${prop}' does not exist`);
            }
            if (["options", "rules", "lexer"].includes(prop)) {
              continue;
            }
            const tokenizerProp = prop;
            const tokenizerFunc = pack.tokenizer[tokenizerProp];
            const prevTokenizer = tokenizer[tokenizerProp];
            tokenizer[tokenizerProp] = (...args2) => {
              let ret = tokenizerFunc.apply(tokenizer, args2);
              if (ret === false) {
                ret = prevTokenizer.apply(tokenizer, args2);
              }
              return ret;
            };
          }
          opts.tokenizer = tokenizer;
        }
        if (pack.hooks) {
          const hooks = this.defaults.hooks || new _Hooks();
          for (const prop in pack.hooks) {
            if (!(prop in hooks)) {
              throw new Error(`hook '${prop}' does not exist`);
            }
            if (["options", "block"].includes(prop)) {
              continue;
            }
            const hooksProp = prop;
            const hooksFunc = pack.hooks[hooksProp];
            const prevHook = hooks[hooksProp];
            if (_Hooks.passThroughHooks.has(prop)) {
              hooks[hooksProp] = (arg) => {
                if (this.defaults.async) {
                  return Promise.resolve(hooksFunc.call(hooks, arg)).then((ret2) => {
                    return prevHook.call(hooks, ret2);
                  });
                }
                const ret = hooksFunc.call(hooks, arg);
                return prevHook.call(hooks, ret);
              };
            } else {
              hooks[hooksProp] = (...args2) => {
                let ret = hooksFunc.apply(hooks, args2);
                if (ret === false) {
                  ret = prevHook.apply(hooks, args2);
                }
                return ret;
              };
            }
          }
          opts.hooks = hooks;
        }
        if (pack.walkTokens) {
          const walkTokens2 = this.defaults.walkTokens;
          const packWalktokens = pack.walkTokens;
          opts.walkTokens = function(token) {
            let values = [];
            values.push(packWalktokens.call(this, token));
            if (walkTokens2) {
              values = values.concat(walkTokens2.call(this, token));
            }
            return values;
          };
        }
        this.defaults = { ...this.defaults, ...opts };
      });
      return this;
    }
    setOptions(opt) {
      this.defaults = { ...this.defaults, ...opt };
      return this;
    }
    lexer(src, options2) {
      return _Lexer.lex(src, options2 ?? this.defaults);
    }
    parser(tokens, options2) {
      return _Parser.parse(tokens, options2 ?? this.defaults);
    }
    parseMarkdown(blockType) {
      const parse2 = (src, options2) => {
        const origOpt = { ...options2 };
        const opt = { ...this.defaults, ...origOpt };
        const throwError = this.onError(!!opt.silent, !!opt.async);
        if (this.defaults.async === true && origOpt.async === false) {
          return throwError(new Error("marked(): The async option was set to true by an extension. Remove async: false from the parse options object to return a Promise."));
        }
        if (typeof src === "undefined" || src === null) {
          return throwError(new Error("marked(): input parameter is undefined or null"));
        }
        if (typeof src !== "string") {
          return throwError(new Error("marked(): input parameter is of type " + Object.prototype.toString.call(src) + ", string expected"));
        }
        if (opt.hooks) {
          opt.hooks.options = opt;
          opt.hooks.block = blockType;
        }
        const lexer2 = opt.hooks ? opt.hooks.provideLexer() : blockType ? _Lexer.lex : _Lexer.lexInline;
        const parser2 = opt.hooks ? opt.hooks.provideParser() : blockType ? _Parser.parse : _Parser.parseInline;
        if (opt.async) {
          return Promise.resolve(opt.hooks ? opt.hooks.preprocess(src) : src).then((src2) => lexer2(src2, opt)).then((tokens) => opt.hooks ? opt.hooks.processAllTokens(tokens) : tokens).then((tokens) => opt.walkTokens ? Promise.all(this.walkTokens(tokens, opt.walkTokens)).then(() => tokens) : tokens).then((tokens) => parser2(tokens, opt)).then((html22) => opt.hooks ? opt.hooks.postprocess(html22) : html22).catch(throwError);
        }
        try {
          if (opt.hooks) {
            src = opt.hooks.preprocess(src);
          }
          let tokens = lexer2(src, opt);
          if (opt.hooks) {
            tokens = opt.hooks.processAllTokens(tokens);
          }
          if (opt.walkTokens) {
            this.walkTokens(tokens, opt.walkTokens);
          }
          let html22 = parser2(tokens, opt);
          if (opt.hooks) {
            html22 = opt.hooks.postprocess(html22);
          }
          return html22;
        } catch (e) {
          return throwError(e);
        }
      };
      return parse2;
    }
    onError(silent, async) {
      return (e) => {
        e.message += "\nPlease report this to https://github.com/markedjs/marked.";
        if (silent) {
          const msg = "<p>An error occurred:</p><pre>" + escape2(e.message + "", true) + "</pre>";
          if (async) {
            return Promise.resolve(msg);
          }
          return msg;
        }
        if (async) {
          return Promise.reject(e);
        }
        throw e;
      };
    }
  };
  var markedInstance = new Marked();
  function marked(src, opt) {
    return markedInstance.parse(src, opt);
  }
  marked.options = marked.setOptions = function(options2) {
    markedInstance.setOptions(options2);
    marked.defaults = markedInstance.defaults;
    changeDefaults(marked.defaults);
    return marked;
  };
  marked.getDefaults = _getDefaults;
  marked.defaults = _defaults;
  marked.use = function(...args) {
    markedInstance.use(...args);
    marked.defaults = markedInstance.defaults;
    changeDefaults(marked.defaults);
    return marked;
  };
  marked.walkTokens = function(tokens, callback) {
    return markedInstance.walkTokens(tokens, callback);
  };
  marked.parseInline = markedInstance.parseInline;
  marked.Parser = _Parser;
  marked.parser = _Parser.parse;
  marked.Renderer = _Renderer;
  marked.TextRenderer = _TextRenderer;
  marked.Lexer = _Lexer;
  marked.lexer = _Lexer.lex;
  marked.Tokenizer = _Tokenizer;
  marked.Hooks = _Hooks;
  marked.parse = marked;
  var options = marked.options;
  var setOptions = marked.setOptions;
  var use = marked.use;
  var walkTokens = marked.walkTokens;
  var parseInline = marked.parseInline;
  var parser = _Parser.parse;
  var lexer = _Lexer.lex;

  // node_modules/dompurify/dist/purify.es.mjs
  function _OverloadYield(e, d) {
    this.v = e, this.k = d;
  }
  function _arrayLikeToArray(r, a) {
    (null == a || a > r.length) && (a = r.length);
    for (var e = 0, n = Array(a); e < a; e++) n[e] = r[e];
    return n;
  }
  function _arrayWithHoles(r) {
    if (Array.isArray(r)) return r;
  }
  function _iterableToArrayLimit(r, l) {
    var t2 = null == r ? null : "undefined" != typeof Symbol && r[Symbol.iterator] || r["@@iterator"];
    if (null != t2) {
      var e, n, i, u, a = [], f = true, o = false;
      try {
        if (i = (t2 = t2.call(r)).next, 0 === l) {
          if (Object(t2) !== t2) return;
          f = false;
        } else for (; !(f = (e = i.call(t2)).done) && (a.push(e.value), a.length !== l); f = true) ;
      } catch (r2) {
        o = true, n = r2;
      } finally {
        try {
          if (!f && null != t2.return && (u = t2.return(), Object(u) !== u)) return;
        } finally {
          if (o) throw n;
        }
      }
      return a;
    }
  }
  function _nonIterableRest() {
    throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
  }
  function _slicedToArray(r, e) {
    return _arrayWithHoles(r) || _iterableToArrayLimit(r, e) || _unsupportedIterableToArray(r, e) || _nonIterableRest();
  }
  function _unsupportedIterableToArray(r, a) {
    if (r) {
      if ("string" == typeof r) return _arrayLikeToArray(r, a);
      var t2 = {}.toString.call(r).slice(8, -1);
      return "Object" === t2 && r.constructor && (t2 = r.constructor.name), "Map" === t2 || "Set" === t2 ? Array.from(r) : "Arguments" === t2 || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(t2) ? _arrayLikeToArray(r, a) : void 0;
    }
  }
  function AsyncGenerator(e) {
    var t2, n;
    function resume(t3, n2) {
      try {
        var r = e[t3](n2), o = r.value, u = o instanceof _OverloadYield;
        Promise.resolve(u ? o.v : o).then(function(n3) {
          if (u) {
            var i = "return" === t3 && o.k ? t3 : "next";
            if (!o.k || n3.done) return resume(i, n3);
            n3 = e[i](n3).value;
          }
          settle(!!r.done, n3);
        }, function(e2) {
          resume("throw", e2);
        });
      } catch (e2) {
        settle(2, e2);
      }
    }
    function settle(e2, r) {
      2 === e2 ? t2.reject(r) : t2.resolve({
        value: r,
        done: e2
      }), (t2 = t2.next) ? resume(t2.key, t2.arg) : n = null;
    }
    this._invoke = function(e2, r) {
      return new Promise(function(o, u) {
        var i = {
          key: e2,
          arg: r,
          resolve: o,
          reject: u,
          next: null
        };
        n ? n = n.next = i : (t2 = n = i, resume(e2, r));
      });
    }, "function" != typeof e.return && (this.return = void 0);
  }
  AsyncGenerator.prototype["function" == typeof Symbol && Symbol.asyncIterator || "@@asyncIterator"] = function() {
    return this;
  }, AsyncGenerator.prototype.next = function(e) {
    return this._invoke("next", e);
  }, AsyncGenerator.prototype.throw = function(e) {
    return this._invoke("throw", e);
  }, AsyncGenerator.prototype.return = function(e) {
    return this._invoke("return", e);
  };
  var entries = Object.entries;
  var objectKeys = Object.keys;
  var setPrototypeOf = Object.setPrototypeOf;
  var isFrozen = Object.isFrozen;
  var getPrototypeOf = Object.getPrototypeOf;
  var getOwnPropertyDescriptor = Object.getOwnPropertyDescriptor;
  var freeze = Object.freeze;
  var seal = Object.seal;
  var create = Object.create;
  var _ref = typeof Reflect !== "undefined" && Reflect;
  var apply = _ref.apply;
  var construct = _ref.construct;
  if (!freeze) freeze = function freeze2(x) {
    return x;
  };
  if (!seal) seal = function seal2(x) {
    return x;
  };
  if (!apply) apply = function apply2(func, thisArg) {
    for (var _len = arguments.length, args = new Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) args[_key - 2] = arguments[_key];
    return func.apply(thisArg, args);
  };
  if (!construct) construct = function construct2(Func) {
    for (var _len2 = arguments.length, args = new Array(_len2 > 1 ? _len2 - 1 : 0), _key2 = 1; _key2 < _len2; _key2++) args[_key2 - 1] = arguments[_key2];
    return new Func(...args);
  };
  var arrayForEach = unapply(Array.prototype.forEach);
  var arrayLastIndexOf = unapply(Array.prototype.lastIndexOf);
  var arrayPop = unapply(Array.prototype.pop);
  var arrayPush = unapply(Array.prototype.push);
  var arraySlice = unapply(Array.prototype.slice);
  var arraySplice = unapply(Array.prototype.splice);
  var arrayIsArray = Array.isArray;
  var setCreate = unconstruct(Set);
  var setAdd = unapply(Set.prototype.add);
  var setHas = unapply(Set.prototype.has);
  var setDelete = unapply(Set.prototype.delete);
  var mapCreate = unconstruct(Map);
  var mapGet = unapply(Map.prototype.get);
  var mapSet = unapply(Map.prototype.set);
  var stringToLowerCase = unapply(String.prototype.toLowerCase);
  var stringToString = unapply(String.prototype.toString);
  var stringMatch = unapply(String.prototype.match);
  var stringReplace = unapply(String.prototype.replace);
  var stringIndexOf = unapply(String.prototype.indexOf);
  var stringTrim = unapply(String.prototype.trim);
  var numberToString = unapply(Number.prototype.toString);
  var booleanToString = unapply(Boolean.prototype.toString);
  var bigintToString = typeof BigInt === "undefined" ? null : unapply(BigInt.prototype.toString);
  var symbolToString = typeof Symbol === "undefined" ? null : unapply(Symbol.prototype.toString);
  var objectHasOwnProperty = unapply(Object.prototype.hasOwnProperty);
  var objectToString = unapply(Object.prototype.toString);
  var regExpTestRaw = RegExp.prototype.test;
  var regExpTest = function regExpTest2(regExp, value) {
    if (regExp instanceof RegExp) regExp.lastIndex = 0;
    return apply(regExpTestRaw, regExp, [value]);
  };
  var typeErrorCreate = unconstruct(TypeError);
  function unapply(func) {
    return function(thisArg) {
      for (var _len3 = arguments.length, args = new Array(_len3 > 1 ? _len3 - 1 : 0), _key3 = 1; _key3 < _len3; _key3++) args[_key3 - 1] = arguments[_key3];
      return apply(func, thisArg, args);
    };
  }
  function unconstruct(Func) {
    return function() {
      for (var _len4 = arguments.length, args = new Array(_len4), _key4 = 0; _key4 < _len4; _key4++) args[_key4] = arguments[_key4];
      return construct(Func, args);
    };
  }
  function addToSet(set2, array) {
    let transformCaseFunc = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : stringToLowerCase;
    if (setPrototypeOf) setPrototypeOf(set2, null);
    if (!arrayIsArray(array)) return set2;
    let l = array.length;
    while (l--) {
      let element = array[l];
      if (typeof element === "string") {
        const lcElement = transformCaseFunc(element);
        if (lcElement !== element) {
          if (!isFrozen(array)) array[l] = lcElement;
          element = lcElement;
        }
      }
      set2[element] = true;
    }
    return set2;
  }
  function cleanArray(array) {
    for (let index = 0; index < array.length; index++) if (!objectHasOwnProperty(array, index)) array[index] = null;
    return array;
  }
  function clone(object) {
    const newObject = create(null);
    for (const _ref2 of entries(object)) {
      var _ref3 = _slicedToArray(_ref2, 2);
      const property = _ref3[0];
      const value = _ref3[1];
      if (objectHasOwnProperty(object, property)) {
        if (arrayIsArray(value)) newObject[property] = cleanArray(value);
        else if (value && typeof value === "object" && value.constructor === Object) newObject[property] = clone(value);
        else newObject[property] = value;
      }
    }
    return newObject;
  }
  function cloneSet(set2) {
    const newSet = create(null);
    const keys = objectKeys(set2);
    let l = keys.length;
    while (l--) newSet[keys[l]] = set2[keys[l]];
    return newSet;
  }
  function stringifyValue(value) {
    switch (typeof value) {
      case "string":
        return value;
      case "number":
        return numberToString(value);
      case "boolean":
        return booleanToString(value);
      case "bigint":
        return bigintToString ? bigintToString(value) : "0";
      case "symbol":
        return symbolToString ? symbolToString(value) : "Symbol()";
      case "undefined":
        return objectToString(value);
      case "function":
      case "object": {
        if (value === null) return objectToString(value);
        const valueAsRecord = value;
        const valueToString = lookupGetter(valueAsRecord, "toString");
        if (typeof valueToString === "function") {
          const stringified = valueToString(valueAsRecord);
          return typeof stringified === "string" ? stringified : objectToString(stringified);
        }
        return objectToString(value);
      }
      default:
        return objectToString(value);
    }
  }
  function lookupGetter(object, prop) {
    while (object !== null) {
      const desc = getOwnPropertyDescriptor(object, prop);
      if (desc) {
        if (desc.get) return unapply(desc.get);
        if (typeof desc.value === "function") return unapply(desc.value);
      }
      object = getPrototypeOf(object);
    }
    function fallbackValue() {
      return null;
    }
    return fallbackValue;
  }
  function lookupAccessor(object, prop) {
    while (object !== null) {
      const desc = getOwnPropertyDescriptor(object, prop);
      if (desc) return desc.get ? unapply(desc.get) : null;
      object = getPrototypeOf(object);
    }
    return null;
  }
  function isRegex(value) {
    if (!value) return false;
    try {
      regExpTest(value, "");
      return true;
    } catch (_unused) {
      return false;
    }
  }
  var html$1 = freeze([
    "a",
    "abbr",
    "acronym",
    "address",
    "area",
    "article",
    "aside",
    "audio",
    "b",
    "bdi",
    "bdo",
    "big",
    "blink",
    "blockquote",
    "body",
    "br",
    "button",
    "canvas",
    "caption",
    "center",
    "cite",
    "code",
    "col",
    "colgroup",
    "content",
    "data",
    "datalist",
    "dd",
    "decorator",
    "del",
    "details",
    "dfn",
    "dialog",
    "dir",
    "div",
    "dl",
    "dt",
    "element",
    "em",
    "fieldset",
    "figcaption",
    "figure",
    "font",
    "footer",
    "form",
    "h1",
    "h2",
    "h3",
    "h4",
    "h5",
    "h6",
    "head",
    "header",
    "hgroup",
    "hr",
    "html",
    "i",
    "img",
    "input",
    "ins",
    "kbd",
    "label",
    "legend",
    "li",
    "main",
    "map",
    "mark",
    "marquee",
    "menu",
    "menuitem",
    "meter",
    "nav",
    "nobr",
    "ol",
    "optgroup",
    "option",
    "output",
    "p",
    "picture",
    "pre",
    "progress",
    "q",
    "rp",
    "rt",
    "ruby",
    "s",
    "samp",
    "search",
    "section",
    "select",
    "shadow",
    "slot",
    "small",
    "source",
    "spacer",
    "span",
    "strike",
    "strong",
    "style",
    "sub",
    "summary",
    "sup",
    "table",
    "tbody",
    "td",
    "template",
    "textarea",
    "tfoot",
    "th",
    "thead",
    "time",
    "tr",
    "track",
    "tt",
    "u",
    "ul",
    "var",
    "video",
    "wbr"
  ]);
  var svg$1 = freeze([
    "svg",
    "a",
    "altglyph",
    "altglyphdef",
    "altglyphitem",
    "animatecolor",
    "animatemotion",
    "animatetransform",
    "circle",
    "clippath",
    "defs",
    "desc",
    "ellipse",
    "filter",
    "font",
    "g",
    "glyph",
    "glyphref",
    "hkern",
    "image",
    "line",
    "lineargradient",
    "marker",
    "mask",
    "metadata",
    "mpath",
    "path",
    "pattern",
    "polygon",
    "polyline",
    "radialgradient",
    "rect",
    "stop",
    "style",
    "switch",
    "symbol",
    "text",
    "textpath",
    "title",
    "tref",
    "tspan",
    "view",
    "vkern"
  ]);
  var svgFilters = freeze([
    "feBlend",
    "feColorMatrix",
    "feComponentTransfer",
    "feComposite",
    "feConvolveMatrix",
    "feDiffuseLighting",
    "feDisplacementMap",
    "feDistantLight",
    "feDropShadow",
    "feFlood",
    "feFuncA",
    "feFuncB",
    "feFuncG",
    "feFuncR",
    "feGaussianBlur",
    "feImage",
    "feMerge",
    "feMergeNode",
    "feMorphology",
    "feOffset",
    "fePointLight",
    "feSpecularLighting",
    "feSpotLight",
    "feTile",
    "feTurbulence"
  ]);
  var svgDisallowed = freeze([
    "animate",
    "color-profile",
    "cursor",
    "discard",
    "font-face",
    "font-face-format",
    "font-face-name",
    "font-face-src",
    "font-face-uri",
    "foreignobject",
    "hatch",
    "hatchpath",
    "mesh",
    "meshgradient",
    "meshpatch",
    "meshrow",
    "missing-glyph",
    "script",
    "set",
    "solidcolor",
    "unknown",
    "use"
  ]);
  var mathMl$1 = freeze([
    "math",
    "menclose",
    "merror",
    "mfenced",
    "mfrac",
    "mglyph",
    "mi",
    "mlabeledtr",
    "mmultiscripts",
    "mn",
    "mo",
    "mover",
    "mpadded",
    "mphantom",
    "mroot",
    "mrow",
    "ms",
    "mspace",
    "msqrt",
    "mstyle",
    "msub",
    "msup",
    "msubsup",
    "mtable",
    "mtd",
    "mtext",
    "mtr",
    "munder",
    "munderover",
    "mprescripts"
  ]);
  var mathMlDisallowed = freeze([
    "maction",
    "maligngroup",
    "malignmark",
    "mlongdiv",
    "mscarries",
    "mscarry",
    "msgroup",
    "mstack",
    "msline",
    "msrow",
    "semantics",
    "annotation",
    "annotation-xml",
    "mprescripts",
    "none"
  ]);
  var text = freeze(["#text"]);
  var html2 = freeze([
    "accept",
    "action",
    "align",
    "alt",
    "autocapitalize",
    "autocomplete",
    "autopictureinpicture",
    "autoplay",
    "background",
    "bgcolor",
    "border",
    "capture",
    "cellpadding",
    "cellspacing",
    "checked",
    "cite",
    "class",
    "clear",
    "color",
    "cols",
    "colspan",
    "command",
    "commandfor",
    "controls",
    "controlslist",
    "coords",
    "crossorigin",
    "datetime",
    "decoding",
    "default",
    "dir",
    "disabled",
    "disablepictureinpicture",
    "disableremoteplayback",
    "download",
    "draggable",
    "enctype",
    "enterkeyhint",
    "exportparts",
    "face",
    "for",
    "headers",
    "height",
    "hidden",
    "high",
    "href",
    "hreflang",
    "id",
    "inert",
    "inputmode",
    "integrity",
    "ismap",
    "kind",
    "label",
    "lang",
    "list",
    "loading",
    "loop",
    "low",
    "max",
    "maxlength",
    "media",
    "method",
    "min",
    "minlength",
    "multiple",
    "muted",
    "name",
    "nonce",
    "noshade",
    "novalidate",
    "nowrap",
    "open",
    "optimum",
    "part",
    "pattern",
    "placeholder",
    "playsinline",
    "popover",
    "popovertarget",
    "popovertargetaction",
    "poster",
    "preload",
    "pubdate",
    "radiogroup",
    "readonly",
    "rel",
    "required",
    "rev",
    "reversed",
    "role",
    "rows",
    "rowspan",
    "spellcheck",
    "scope",
    "selected",
    "shape",
    "size",
    "sizes",
    "slot",
    "span",
    "srclang",
    "start",
    "src",
    "srcset",
    "step",
    "style",
    "summary",
    "tabindex",
    "title",
    "translate",
    "type",
    "usemap",
    "valign",
    "value",
    "width",
    "wrap",
    "xmlns"
  ]);
  var svg = freeze([
    "accent-height",
    "accumulate",
    "additive",
    "alignment-baseline",
    "amplitude",
    "ascent",
    "attributename",
    "attributetype",
    "azimuth",
    "basefrequency",
    "baseline-shift",
    "begin",
    "bias",
    "by",
    "class",
    "clip",
    "clippathunits",
    "clip-path",
    "clip-rule",
    "color",
    "color-interpolation",
    "color-interpolation-filters",
    "color-profile",
    "color-rendering",
    "cx",
    "cy",
    "d",
    "dx",
    "dy",
    "diffuseconstant",
    "direction",
    "display",
    "divisor",
    "dominant-baseline",
    "dur",
    "edgemode",
    "elevation",
    "end",
    "exponent",
    "fill",
    "fill-opacity",
    "fill-rule",
    "filter",
    "filterunits",
    "flood-color",
    "flood-opacity",
    "font-family",
    "font-size",
    "font-size-adjust",
    "font-stretch",
    "font-style",
    "font-variant",
    "font-weight",
    "fx",
    "fy",
    "g1",
    "g2",
    "glyph-name",
    "glyphref",
    "gradientunits",
    "gradienttransform",
    "height",
    "href",
    "id",
    "image-rendering",
    "in",
    "in2",
    "intercept",
    "k",
    "k1",
    "k2",
    "k3",
    "k4",
    "kerning",
    "keypoints",
    "keysplines",
    "keytimes",
    "lang",
    "lengthadjust",
    "letter-spacing",
    "kernelmatrix",
    "kernelunitlength",
    "lighting-color",
    "local",
    "marker-end",
    "marker-mid",
    "marker-start",
    "markerheight",
    "markerunits",
    "markerwidth",
    "maskcontentunits",
    "maskunits",
    "max",
    "mask",
    "mask-type",
    "media",
    "method",
    "mode",
    "min",
    "name",
    "numoctaves",
    "offset",
    "operator",
    "opacity",
    "order",
    "orient",
    "orientation",
    "origin",
    "overflow",
    "paint-order",
    "path",
    "pathlength",
    "patterncontentunits",
    "patterntransform",
    "patternunits",
    "pointer-events",
    "points",
    "preservealpha",
    "preserveaspectratio",
    "primitiveunits",
    "r",
    "rx",
    "ry",
    "radius",
    "refx",
    "refy",
    "repeatcount",
    "repeatdur",
    "restart",
    "result",
    "rotate",
    "scale",
    "seed",
    "shape-rendering",
    "slope",
    "specularconstant",
    "specularexponent",
    "spreadmethod",
    "startoffset",
    "stddeviation",
    "stitchtiles",
    "stop-color",
    "stop-opacity",
    "stroke-dasharray",
    "stroke-dashoffset",
    "stroke-linecap",
    "stroke-linejoin",
    "stroke-miterlimit",
    "stroke-opacity",
    "stroke",
    "stroke-width",
    "style",
    "surfacescale",
    "systemlanguage",
    "tabindex",
    "tablevalues",
    "targetx",
    "targety",
    "transform",
    "transform-origin",
    "text-anchor",
    "text-decoration",
    "text-orientation",
    "text-rendering",
    "textlength",
    "type",
    "u1",
    "u2",
    "unicode",
    "values",
    "vector-effect",
    "viewbox",
    "visibility",
    "version",
    "vert-adv-y",
    "vert-origin-x",
    "vert-origin-y",
    "width",
    "word-spacing",
    "wrap",
    "writing-mode",
    "xchannelselector",
    "ychannelselector",
    "x",
    "x1",
    "x2",
    "xmlns",
    "y",
    "y1",
    "y2",
    "z",
    "zoomandpan"
  ]);
  var mathMl = freeze([
    "accent",
    "accentunder",
    "align",
    "bevelled",
    "close",
    "columnalign",
    "columnlines",
    "columnspacing",
    "columnspan",
    "denomalign",
    "depth",
    "dir",
    "display",
    "displaystyle",
    "encoding",
    "fence",
    "frame",
    "height",
    "href",
    "id",
    "largeop",
    "length",
    "linethickness",
    "lquote",
    "lspace",
    "mathbackground",
    "mathcolor",
    "mathsize",
    "mathvariant",
    "maxsize",
    "minsize",
    "movablelimits",
    "notation",
    "numalign",
    "open",
    "rowalign",
    "rowlines",
    "rowspacing",
    "rowspan",
    "rspace",
    "rquote",
    "scriptlevel",
    "scriptminsize",
    "scriptsizemultiplier",
    "selection",
    "separator",
    "separators",
    "stretchy",
    "subscriptshift",
    "supscriptshift",
    "symmetric",
    "voffset",
    "width",
    "xmlns"
  ]);
  var xml = freeze([
    "xlink:href",
    "xml:id",
    "xlink:title",
    "xml:space",
    "xmlns:xlink"
  ]);
  var MUSTACHE_EXPR = seal(/{{[\w\W]*|^[\w\W]*}}/g);
  var ERB_EXPR = seal(/<%[\w\W]*|^[\w\W]*%>/g);
  var TMPLIT_EXPR = seal(/\${[\w\W]*/g);
  var TMPL_MARKER = seal(/{{|}}|\${|<%|%>/);
  var NON_WHITESPACE = seal(/[^\t\n\f\r ]/);
  var NUL = seal(/\u0000/g);
  var DATA_ATTR = seal(/^data-[\-\w.\u00B7-\uFFFF]+$/);
  var ARIA_ATTR = seal(/^aria-[\-\w]+$/);
  var IS_ALLOWED_URI = seal(/^(?:(?:(?:f|ht)tps?|mailto|tel|callto|sms|cid|xmpp|matrix):|[^a-z]|[a-z+.\-]+(?:[^a-z+.\-:]|$))/i);
  var IS_SCRIPT_OR_DATA = seal(/^(?:\w+script|data):/i);
  var ATTR_WHITESPACE = seal(/[\u0000-\u0020\u00A0\u1680\u180E\u2000-\u2029\u205F\u3000]/g);
  var DOCTYPE_NAME = seal(/^html$/i);
  var CUSTOM_ELEMENT = seal(/^[a-z][.\w]*(-[.\w]+)+$/i);
  var ELEMENT_MARKUP_PROBE = seal(/<[/\w!]/g);
  var COMMENT_MARKUP_PROBE = seal(/<[/\w]/g);
  var COMMENT_UNSAFE = seal(/<[/\w]|^-?>|--!?>|<!-$/g);
  var FALLBACK_TAG_CLOSE = seal(/<\/no(script|embed|frames)/i);
  var SELF_CLOSING_TAG = seal(/\/>/i);
  var ATTR_VALUE_BREAKOUT = seal(/((--!?|])>)|<\/(style|script|title|xmp|textarea|noscript|iframe|noembed|noframes)/i);
  var LEADING_WHITESPACE = seal(/^[\r\n\t ]+/);
  var NODE_TYPE = {
    element: 1,
    text: 3,
    processingInstruction: 7,
    comment: 8,
    document: 9,
    documentFragment: 11
  };
  var LITERAL_TEXT_ELEMENTS = freeze(addToSet({}, [
    "style",
    "script",
    "xmp",
    "iframe",
    "noembed",
    "noframes",
    "plaintext",
    "noscript"
  ]));
  var ESCAPABLE_TEXT_ELEMENTS = freeze(addToSet({}, ["textarea", "title"]));
  var VOID_ELEMENTS = freeze(addToSet({}, [
    "area",
    "base",
    "basefont",
    "bgsound",
    "br",
    "col",
    "embed",
    "frame",
    "hr",
    "img",
    "input",
    "keygen",
    "link",
    "meta",
    "param",
    "source",
    "track",
    "wbr"
  ]));
  var getGlobal = function getGlobal2() {
    return typeof window === "undefined" ? null : window;
  };
  var _createTrustedTypesPolicy = function _createTrustedTypesPolicy2(trustedTypes, purifyHostElement) {
    if (typeof trustedTypes !== "object" || typeof trustedTypes.createPolicy !== "function") return null;
    let suffix = null;
    const ATTR_NAME = "data-tt-policy-suffix";
    if (purifyHostElement && purifyHostElement.hasAttribute(ATTR_NAME)) suffix = purifyHostElement.getAttribute(ATTR_NAME);
    const policyName = "dompurify" + (suffix ? "#" + suffix : "");
    try {
      return trustedTypes.createPolicy(policyName, {
        createHTML(html3) {
          return html3;
        },
        createScriptURL(scriptUrl) {
          return scriptUrl;
        }
      });
    } catch (_) {
      console.warn("TrustedTypes policy " + policyName + " could not be created.");
      return null;
    }
  };
  var _createHooksMap = function _createHooksMap2() {
    return {
      afterSanitizeAttributes: [],
      afterSanitizeElements: [],
      afterSanitizeShadowDOM: [],
      beforeSanitizeAttributes: [],
      beforeSanitizeElements: [],
      beforeSanitizeShadowDOM: [],
      uponSanitizeAttribute: [],
      uponSanitizeElement: [],
      uponSanitizeShadowNode: []
    };
  };
  var _resolveSetOption = function _resolveSetOption2(cfg, key, fallback, options2) {
    return objectHasOwnProperty(cfg, key) && arrayIsArray(cfg[key]) ? addToSet(options2.base ? cloneSet(options2.base) : {}, cfg[key], options2.transform) : fallback;
  };
  var _resolveObjectOption = function _resolveObjectOption2(cfg, key, fallback) {
    const value = objectHasOwnProperty(cfg, key) ? cfg[key] : void 0;
    return value && typeof value === "object" ? clone(value) : fallback;
  };
  function createDOMPurify() {
    let window2 = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : getGlobal();
    const DOMPurify = (root) => createDOMPurify(root);
    DOMPurify.version = "3.5.0";
    DOMPurify.removed = [];
    if (!window2 || !window2.document || window2.document.nodeType !== NODE_TYPE.document || !window2.Element) {
      DOMPurify.isSupported = false;
      return DOMPurify;
    }
    let document2 = window2.document;
    const originalDocument = document2;
    const currentScript = originalDocument.currentScript;
    const HTMLTemplateElement = window2.HTMLTemplateElement, Node = window2.Node, Element = window2.Element, NodeFilter = window2.NodeFilter, DOMParser = window2.DOMParser, trustedTypes = window2.trustedTypes;
    const ElementPrototype = Element.prototype;
    const cloneNode = lookupGetter(ElementPrototype, "cloneNode");
    const remove = lookupGetter(ElementPrototype, "remove");
    const removeChild = lookupGetter(ElementPrototype, "removeChild");
    const removeAttributeNode = lookupGetter(ElementPrototype, "removeAttributeNode");
    const getNextSibling = lookupGetter(ElementPrototype, "nextSibling");
    const getChildNodes = lookupGetter(ElementPrototype, "childNodes");
    const getParentNode = lookupGetter(ElementPrototype, "parentNode");
    const getShadowRoot = lookupGetter(ElementPrototype, "shadowRoot");
    const getAttributes = lookupGetter(ElementPrototype, "attributes");
    const getNodeType = Node && Node.prototype ? lookupGetter(Node.prototype, "nodeType") : null;
    const getNodeName = Node && Node.prototype ? lookupGetter(Node.prototype, "nodeName") : null;
    const getOwnerDocument = Node && Node.prototype ? lookupGetter(Node.prototype, "ownerDocument") : null;
    const _readNodeType = function _readNodeType2(node) {
      return getNodeType ? getNodeType(node) : node.nodeType;
    };
    const _readNodeName = function _readNodeName2(node) {
      return getNodeName ? getNodeName(node) : node.nodeName;
    };
    const getNamespaceURI = lookupAccessor(ElementPrototype, "namespaceURI");
    const _readNamespaceURI = function _readNamespaceURI2(element) {
      return getNamespaceURI ? getNamespaceURI(element) : element.namespaceURI;
    };
    const _readAttributes = function _readAttributes2(node) {
      return _readNodeType(node) === NODE_TYPE.element ? getAttributes(node) : null;
    };
    if (typeof HTMLTemplateElement === "function") {
      const template = document2.createElement("template");
      if (template.content && template.content.ownerDocument) document2 = template.content.ownerDocument;
    }
    let trustedTypesPolicy;
    let emptyHTML = "";
    let defaultTrustedTypesPolicy;
    let defaultTrustedTypesPolicyResolved = false;
    let IN_TRUSTED_TYPES_POLICY = 0;
    const _assertNotInTrustedTypesPolicy = function _assertNotInTrustedTypesPolicy2() {
      if (IN_TRUSTED_TYPES_POLICY > 0) throw typeErrorCreate('A configured TRUSTED_TYPES_POLICY callback (createHTML or createScriptURL) must not call DOMPurify.sanitize, as that causes infinite recursion. Do not pass a policy whose callbacks wrap DOMPurify as TRUSTED_TYPES_POLICY; see the "DOMPurify and Trusted Types" section of the README.');
    };
    const _createTrustedHTML = function _createTrustedHTML2(html3) {
      _assertNotInTrustedTypesPolicy();
      IN_TRUSTED_TYPES_POLICY++;
      try {
        return trustedTypesPolicy.createHTML(html3);
      } finally {
        IN_TRUSTED_TYPES_POLICY--;
      }
    };
    const _signOutput = function _signOutput2(html3) {
      return trustedTypesPolicy && RETURN_TRUSTED_TYPE ? _createTrustedHTML(html3) : html3;
    };
    const _createTrustedScriptURL = function _createTrustedScriptURL2(scriptUrl) {
      _assertNotInTrustedTypesPolicy();
      IN_TRUSTED_TYPES_POLICY++;
      try {
        return trustedTypesPolicy.createScriptURL(scriptUrl);
      } finally {
        IN_TRUSTED_TYPES_POLICY--;
      }
    };
    const _getDefaultTrustedTypesPolicy = function _getDefaultTrustedTypesPolicy2() {
      if (!defaultTrustedTypesPolicyResolved) {
        defaultTrustedTypesPolicy = _createTrustedTypesPolicy(trustedTypes, currentScript);
        defaultTrustedTypesPolicyResolved = true;
      }
      return defaultTrustedTypesPolicy;
    };
    const _document = document2, implementation = _document.implementation, createNodeIterator = _document.createNodeIterator, createDocumentFragment = _document.createDocumentFragment, getElementsByTagName = _document.getElementsByTagName;
    const importNode = originalDocument.importNode;
    let hooks = _createHooksMap();
    DOMPurify.isSupported = typeof entries === "function" && typeof getParentNode === "function" && implementation && implementation.createHTMLDocument !== void 0;
    const MUSTACHE_EXPR$1 = MUSTACHE_EXPR, ERB_EXPR$1 = ERB_EXPR, TMPLIT_EXPR$1 = TMPLIT_EXPR, DATA_ATTR$1 = DATA_ATTR, ARIA_ATTR$1 = ARIA_ATTR, IS_SCRIPT_OR_DATA$1 = IS_SCRIPT_OR_DATA, ATTR_WHITESPACE$1 = ATTR_WHITESPACE, CUSTOM_ELEMENT$1 = CUSTOM_ELEMENT;
    let IS_ALLOWED_URI$1 = IS_ALLOWED_URI;
    let ALLOWED_TAGS = null;
    const DEFAULT_ALLOWED_TAGS = addToSet({}, [
      ...html$1,
      ...svg$1,
      ...svgFilters,
      ...mathMl$1,
      ...text
    ]);
    let ALLOWED_ATTR = null;
    const DEFAULT_ALLOWED_ATTR = addToSet({}, [
      ...html2,
      ...svg,
      ...mathMl,
      ...xml
    ]);
    let CUSTOM_ELEMENT_HANDLING = null;
    let FORBID_TAGS = null;
    let FORBID_ATTR = null;
    const EXTRA_ELEMENT_HANDLING = Object.seal(create(null, {
      tagCheck: {
        writable: true,
        configurable: false,
        enumerable: true,
        value: null
      },
      attributeCheck: {
        writable: true,
        configurable: false,
        enumerable: true,
        value: null
      }
    }));
    let ALLOW_ARIA_ATTR = true;
    let ALLOW_DATA_ATTR = true;
    let ALLOW_UNKNOWN_PROTOCOLS = false;
    let ALLOW_SELF_CLOSE_IN_ATTR = true;
    let SAFE_FOR_TEMPLATES = false;
    let SAFE_FOR_XML = true;
    let WHOLE_DOCUMENT = false;
    let SET_CONFIG = false;
    let SET_CONFIG_ALLOWED_TAGS = null;
    let SET_CONFIG_ALLOWED_ATTR = null;
    let FORCE_BODY = false;
    let RETURN_DOM = false;
    let RETURN_DOM_FRAGMENT = false;
    let RETURN_TRUSTED_TYPE = false;
    let SANITIZE_DOM = true;
    let SANITIZE_NAMED_PROPS = false;
    const SANITIZE_NAMED_PROPS_PREFIX = "user-content-";
    let KEEP_CONTENT = true;
    let IN_PLACE = false;
    let IN_PLACE_VISITED = null;
    let IN_PLACE_RETURNS = null;
    let USE_PROFILES = false;
    let FORBID_CONTENTS = null;
    const DEFAULT_FORBID_CONTENTS = addToSet({}, [
      "annotation-xml",
      "audio",
      "colgroup",
      "desc",
      "foreignobject",
      "head",
      "iframe",
      "math",
      "mi",
      "mn",
      "mo",
      "ms",
      "mtext",
      "noembed",
      "noframes",
      "noscript",
      "plaintext",
      "script",
      "selectedcontent",
      "style",
      "svg",
      "template",
      "thead",
      "title",
      "video",
      "xmp"
    ]);
    let DATA_URI_TAGS = null;
    const DEFAULT_DATA_URI_TAGS = addToSet({}, [
      "audio",
      "video",
      "img",
      "source",
      "image",
      "track"
    ]);
    let URI_SAFE_ATTRIBUTES = null;
    const DEFAULT_URI_SAFE_ATTRIBUTES = addToSet({}, [
      "alt",
      "class",
      "for",
      "id",
      "label",
      "name",
      "pattern",
      "placeholder",
      "role",
      "summary",
      "title",
      "value",
      "style",
      "xmlns"
    ]);
    const MATHML_NAMESPACE = "http://www.w3.org/1998/Math/MathML";
    const SVG_NAMESPACE = "http://www.w3.org/2000/svg";
    const HTML_NAMESPACE = "http://www.w3.org/1999/xhtml";
    let NAMESPACE = HTML_NAMESPACE;
    let IS_EMPTY_INPUT = false;
    let ALLOWED_NAMESPACES = null;
    const DEFAULT_ALLOWED_NAMESPACES = addToSet({}, [
      MATHML_NAMESPACE,
      SVG_NAMESPACE,
      HTML_NAMESPACE
    ], stringToString);
    const DEFAULT_MATHML_TEXT_INTEGRATION_POINTS = freeze(addToSet({}, [
      "mi",
      "mo",
      "mn",
      "ms",
      "mtext"
    ]));
    let MATHML_TEXT_INTEGRATION_POINTS = null;
    const DEFAULT_HTML_INTEGRATION_POINTS = freeze(addToSet({}, ["annotation-xml"]));
    let HTML_INTEGRATION_POINTS = null;
    const DEFAULT_CUSTOM_ELEMENT_HANDLING = freeze(create(null));
    const COMMON_SVG_AND_HTML_ELEMENTS = addToSet({}, [
      "title",
      "style",
      "font",
      "a",
      "script"
    ]);
    let PARSER_MEDIA_TYPE = null;
    const SUPPORTED_PARSER_MEDIA_TYPES = ["application/xhtml+xml", "text/html"];
    const DEFAULT_PARSER_MEDIA_TYPE = "text/html";
    let transformCaseFunc = null;
    let CONFIG = null;
    let ACTIVE_CONFIG = null;
    const formElement = document2.createElement("form");
    const isRegexOrFunction = function isRegexOrFunction2(testValue) {
      return testValue instanceof RegExp || testValue instanceof Function;
    };
    const _parseConfig = function _parseConfig2() {
      let cfg = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
      if (!cfg || typeof cfg !== "object") cfg = {};
      cfg = clone(cfg);
      PARSER_MEDIA_TYPE = SUPPORTED_PARSER_MEDIA_TYPES.indexOf(cfg.PARSER_MEDIA_TYPE) === -1 ? DEFAULT_PARSER_MEDIA_TYPE : cfg.PARSER_MEDIA_TYPE;
      transformCaseFunc = PARSER_MEDIA_TYPE === "application/xhtml+xml" ? stringToString : stringToLowerCase;
      ALLOWED_TAGS = _resolveSetOption(cfg, "ALLOWED_TAGS", DEFAULT_ALLOWED_TAGS, { transform: transformCaseFunc });
      ALLOWED_ATTR = _resolveSetOption(cfg, "ALLOWED_ATTR", DEFAULT_ALLOWED_ATTR, { transform: transformCaseFunc });
      ALLOWED_NAMESPACES = _resolveSetOption(cfg, "ALLOWED_NAMESPACES", DEFAULT_ALLOWED_NAMESPACES, { transform: stringToString });
      URI_SAFE_ATTRIBUTES = _resolveSetOption(cfg, "ADD_URI_SAFE_ATTR", DEFAULT_URI_SAFE_ATTRIBUTES, {
        transform: transformCaseFunc,
        base: DEFAULT_URI_SAFE_ATTRIBUTES
      });
      DATA_URI_TAGS = _resolveSetOption(cfg, "ADD_DATA_URI_TAGS", DEFAULT_DATA_URI_TAGS, {
        transform: transformCaseFunc,
        base: DEFAULT_DATA_URI_TAGS
      });
      FORBID_CONTENTS = _resolveSetOption(cfg, "FORBID_CONTENTS", DEFAULT_FORBID_CONTENTS, { transform: transformCaseFunc });
      FORBID_TAGS = _resolveSetOption(cfg, "FORBID_TAGS", create(null), { transform: transformCaseFunc });
      FORBID_ATTR = _resolveSetOption(cfg, "FORBID_ATTR", create(null), { transform: transformCaseFunc });
      USE_PROFILES = objectHasOwnProperty(cfg, "USE_PROFILES") ? cfg.USE_PROFILES && typeof cfg.USE_PROFILES === "object" ? clone(cfg.USE_PROFILES) : cfg.USE_PROFILES : false;
      ALLOW_ARIA_ATTR = cfg.ALLOW_ARIA_ATTR !== false;
      ALLOW_DATA_ATTR = cfg.ALLOW_DATA_ATTR !== false;
      ALLOW_UNKNOWN_PROTOCOLS = cfg.ALLOW_UNKNOWN_PROTOCOLS || false;
      ALLOW_SELF_CLOSE_IN_ATTR = cfg.ALLOW_SELF_CLOSE_IN_ATTR !== false;
      SAFE_FOR_TEMPLATES = cfg.SAFE_FOR_TEMPLATES || false;
      SAFE_FOR_XML = cfg.SAFE_FOR_XML !== false;
      WHOLE_DOCUMENT = cfg.WHOLE_DOCUMENT || false;
      RETURN_DOM = cfg.RETURN_DOM || false;
      RETURN_DOM_FRAGMENT = cfg.RETURN_DOM_FRAGMENT || false;
      RETURN_TRUSTED_TYPE = cfg.RETURN_TRUSTED_TYPE || false;
      FORCE_BODY = cfg.FORCE_BODY || false;
      SANITIZE_DOM = cfg.SANITIZE_DOM !== false;
      SANITIZE_NAMED_PROPS = cfg.SANITIZE_NAMED_PROPS || false;
      KEEP_CONTENT = cfg.KEEP_CONTENT !== false;
      IN_PLACE = cfg.IN_PLACE || false;
      IS_ALLOWED_URI$1 = isRegex(cfg.ALLOWED_URI_REGEXP) ? cfg.ALLOWED_URI_REGEXP : IS_ALLOWED_URI;
      NAMESPACE = typeof cfg.NAMESPACE === "string" ? cfg.NAMESPACE : HTML_NAMESPACE;
      MATHML_TEXT_INTEGRATION_POINTS = _resolveObjectOption(cfg, "MATHML_TEXT_INTEGRATION_POINTS", DEFAULT_MATHML_TEXT_INTEGRATION_POINTS);
      HTML_INTEGRATION_POINTS = _resolveObjectOption(cfg, "HTML_INTEGRATION_POINTS", DEFAULT_HTML_INTEGRATION_POINTS);
      const customElementHandling = _resolveObjectOption(cfg, "CUSTOM_ELEMENT_HANDLING", DEFAULT_CUSTOM_ELEMENT_HANDLING);
      CUSTOM_ELEMENT_HANDLING = create(null);
      if (objectHasOwnProperty(customElementHandling, "tagNameCheck") && isRegexOrFunction(customElementHandling.tagNameCheck)) CUSTOM_ELEMENT_HANDLING.tagNameCheck = customElementHandling.tagNameCheck;
      if (objectHasOwnProperty(customElementHandling, "attributeNameCheck") && isRegexOrFunction(customElementHandling.attributeNameCheck)) CUSTOM_ELEMENT_HANDLING.attributeNameCheck = customElementHandling.attributeNameCheck;
      if (objectHasOwnProperty(customElementHandling, "allowCustomizedBuiltInElements") && typeof customElementHandling.allowCustomizedBuiltInElements === "boolean") CUSTOM_ELEMENT_HANDLING.allowCustomizedBuiltInElements = customElementHandling.allowCustomizedBuiltInElements;
      seal(CUSTOM_ELEMENT_HANDLING);
      if (SAFE_FOR_TEMPLATES) ALLOW_DATA_ATTR = false;
      if (RETURN_DOM_FRAGMENT) RETURN_DOM = true;
      if (USE_PROFILES) {
        if ((objectHasOwnProperty(cfg, "ALLOWED_TAGS") || objectHasOwnProperty(cfg, "ALLOWED_ATTR")) && typeof console !== "undefined" && typeof console.warn === "function") console.warn("DOMPurify: USE_PROFILES overrides ALLOWED_TAGS and ALLOWED_ATTR. Do not use these options together.");
        ALLOWED_TAGS = addToSet({}, text);
        ALLOWED_ATTR = create(null);
        if (USE_PROFILES.html === true) {
          addToSet(ALLOWED_TAGS, html$1);
          addToSet(ALLOWED_ATTR, html2);
        }
        if (USE_PROFILES.svg === true) {
          addToSet(ALLOWED_TAGS, svg$1);
          addToSet(ALLOWED_ATTR, svg);
          addToSet(ALLOWED_ATTR, xml);
        }
        if (USE_PROFILES.svgFilters === true) {
          addToSet(ALLOWED_TAGS, svgFilters);
          addToSet(ALLOWED_ATTR, svg);
          addToSet(ALLOWED_ATTR, xml);
        }
        if (USE_PROFILES.mathMl === true) {
          addToSet(ALLOWED_TAGS, mathMl$1);
          addToSet(ALLOWED_ATTR, mathMl);
          addToSet(ALLOWED_ATTR, xml);
        }
      }
      EXTRA_ELEMENT_HANDLING.tagCheck = null;
      EXTRA_ELEMENT_HANDLING.attributeCheck = null;
      if (objectHasOwnProperty(cfg, "ADD_TAGS")) {
        if (typeof cfg.ADD_TAGS === "function") EXTRA_ELEMENT_HANDLING.tagCheck = cfg.ADD_TAGS;
        else if (arrayIsArray(cfg.ADD_TAGS)) {
          if (ALLOWED_TAGS === DEFAULT_ALLOWED_TAGS) ALLOWED_TAGS = cloneSet(ALLOWED_TAGS);
          addToSet(ALLOWED_TAGS, cfg.ADD_TAGS, transformCaseFunc);
        }
      }
      if (objectHasOwnProperty(cfg, "ADD_ATTR")) {
        if (typeof cfg.ADD_ATTR === "function") EXTRA_ELEMENT_HANDLING.attributeCheck = cfg.ADD_ATTR;
        else if (arrayIsArray(cfg.ADD_ATTR)) {
          if (ALLOWED_ATTR === DEFAULT_ALLOWED_ATTR) ALLOWED_ATTR = cloneSet(ALLOWED_ATTR);
          addToSet(ALLOWED_ATTR, cfg.ADD_ATTR, transformCaseFunc);
        }
      }
      if (objectHasOwnProperty(cfg, "ADD_FORBID_CONTENTS") && arrayIsArray(cfg.ADD_FORBID_CONTENTS)) {
        if (FORBID_CONTENTS === DEFAULT_FORBID_CONTENTS) FORBID_CONTENTS = cloneSet(FORBID_CONTENTS);
        addToSet(FORBID_CONTENTS, cfg.ADD_FORBID_CONTENTS, transformCaseFunc);
      }
      if (KEEP_CONTENT) ALLOWED_TAGS["#text"] = true;
      if (WHOLE_DOCUMENT) addToSet(ALLOWED_TAGS, [
        "html",
        "head",
        "body"
      ]);
      if (ALLOWED_TAGS.table) {
        ALLOWED_TAGS.tbody = true;
        delete FORBID_TAGS.tbody;
      }
      if (cfg.TRUSTED_TYPES_POLICY) {
        if (typeof cfg.TRUSTED_TYPES_POLICY.createHTML !== "function") throw typeErrorCreate('TRUSTED_TYPES_POLICY configuration option must provide a "createHTML" hook.');
        if (typeof cfg.TRUSTED_TYPES_POLICY.createScriptURL !== "function") throw typeErrorCreate('TRUSTED_TYPES_POLICY configuration option must provide a "createScriptURL" hook.');
        const previousTrustedTypesPolicy = trustedTypesPolicy;
        trustedTypesPolicy = cfg.TRUSTED_TYPES_POLICY;
        try {
          emptyHTML = _createTrustedHTML("");
        } catch (error) {
          trustedTypesPolicy = previousTrustedTypesPolicy;
          throw error;
        }
      } else if (cfg.TRUSTED_TYPES_POLICY === null) {
        trustedTypesPolicy = void 0;
        emptyHTML = "";
      } else {
        if (trustedTypesPolicy === void 0) trustedTypesPolicy = _getDefaultTrustedTypesPolicy();
        if (trustedTypesPolicy && typeof emptyHTML === "string") emptyHTML = _createTrustedHTML("");
      }
      CONFIG = freeze(cfg);
    };
    const ALL_SVG_TAGS = addToSet({}, [
      ...svg$1,
      ...svgFilters,
      ...svgDisallowed
    ]);
    const ALL_MATHML_TAGS = addToSet({}, [...mathMl$1, ...mathMlDisallowed]);
    const _checkSvgNamespace = function _checkSvgNamespace2(tagName, parentNamespace, parentTagName) {
      if (parentNamespace === HTML_NAMESPACE) return tagName === "svg";
      if (parentNamespace === MATHML_NAMESPACE) return tagName === "svg" && (parentTagName === "annotation-xml" || MATHML_TEXT_INTEGRATION_POINTS[parentTagName]);
      return Boolean(ALL_SVG_TAGS[tagName]);
    };
    const _checkMathMlNamespace = function _checkMathMlNamespace2(tagName, parentNamespace, parentTagName) {
      if (parentNamespace === HTML_NAMESPACE) return tagName === "math";
      if (parentNamespace === SVG_NAMESPACE) return tagName === "math" && HTML_INTEGRATION_POINTS[parentTagName];
      return Boolean(ALL_MATHML_TAGS[tagName]);
    };
    const _checkHtmlNamespace = function _checkHtmlNamespace2(tagName, parentNamespace, parentTagName) {
      if (parentNamespace === SVG_NAMESPACE && !HTML_INTEGRATION_POINTS[parentTagName]) return false;
      if (parentNamespace === MATHML_NAMESPACE && !MATHML_TEXT_INTEGRATION_POINTS[parentTagName]) return false;
      return !ALL_MATHML_TAGS[tagName] && (COMMON_SVG_AND_HTML_ELEMENTS[tagName] || !ALL_SVG_TAGS[tagName]);
    };
    const _checkValidNamespace = function _checkValidNamespace2(element) {
      const elementNamespace = _readNamespaceURI(element);
      const tagName = stringToLowerCase(_readNodeName(element));
      const parentNode = getParentNode(element);
      let parentNamespace = NAMESPACE;
      let parentTagName = "template";
      if (parentNode && _readNodeType(parentNode) === NODE_TYPE.element) {
        parentNamespace = _readNamespaceURI(parentNode);
        parentTagName = stringToLowerCase(_readNodeName(parentNode));
      }
      if (!ALLOWED_NAMESPACES[elementNamespace]) return false;
      if (elementNamespace === SVG_NAMESPACE) return _checkSvgNamespace(tagName, parentNamespace, parentTagName);
      if (elementNamespace === MATHML_NAMESPACE) return _checkMathMlNamespace(tagName, parentNamespace, parentTagName);
      if (elementNamespace === HTML_NAMESPACE) return _checkHtmlNamespace(tagName, parentNamespace, parentTagName);
      return PARSER_MEDIA_TYPE === "application/xhtml+xml";
    };
    const _forceRemove = function _forceRemove2(node) {
      arrayPush(DOMPurify.removed, { element: node });
      _unmarkVisited(node);
      try {
        getParentNode(node).removeChild(node);
      } catch (_) {
        remove(node);
        if (!getParentNode(node)) throw typeErrorCreate("a node selected for removal could not be detached from its tree and cannot be safely returned; refusing to sanitize in place");
      }
    };
    const _stripAttributeNode = function _stripAttributeNode2(element, attribute, name) {
      try {
        removeAttributeNode(element, attribute);
      } catch (_) {
        try {
          element.removeAttribute(name);
        } catch (_2) {
        }
      }
    };
    const _neutralizeRoot = function _neutralizeRoot2(root) {
      _neutralizeSubtree(root);
      const childNodes = getChildNodes(root);
      if (childNodes) arrayForEach(arraySlice(childNodes), (child) => {
        try {
          removeChild(root, child);
        } catch (_) {
        }
      });
      if (_readNodeType(root) === NODE_TYPE.element) _stripAttributes(root, false);
    };
    const _removeAttribute = function _removeAttribute2(name, element, attr) {
      if (!attr) try {
        attr = element.getAttributeNode(name);
      } catch (_) {
        attr = null;
      }
      arrayPush(DOMPurify.removed, {
        attribute: attr || null,
        from: element
      });
      if (attr) _stripAttributeNode(element, attr, name);
      else try {
        element.removeAttribute(name);
      } catch (_) {
      }
      if (name === "is") {
        if (RETURN_DOM || RETURN_DOM_FRAGMENT) try {
          _forceRemove(element);
        } catch (_) {
        }
        else try {
          element.setAttribute(name, "");
        } catch (_) {
        }
      }
    };
    const _stripAttributes = function _stripAttributes2(element, keepAllowed) {
      const attributes = getAttributes(element);
      if (!attributes) return;
      for (let i = attributes.length - 1; i >= 0; --i) {
        const attribute = attributes[i];
        const name = attribute && attribute.name;
        if (typeof name === "string" && !(keepAllowed && ALLOWED_ATTR[transformCaseFunc(name)])) _stripAttributeNode(element, attribute, name);
      }
    };
    const _walkTree = function _walkTree2(root, visit) {
      let intoShadow = arguments.length > 2 && arguments[2] !== void 0 ? arguments[2] : false;
      const stack = [root];
      while (stack.length > 0) {
        const node = stack.pop();
        if (visit(node) === false) continue;
        if (intoShadow && _readNodeType(node) === NODE_TYPE.element) {
          const shadowRoot = getShadowRoot(node);
          if (_isDocumentFragment(shadowRoot)) stack.push(shadowRoot);
        }
        const childNodes = getChildNodes(node);
        if (childNodes) for (let i = childNodes.length - 1; i >= 0; --i) stack.push(childNodes[i]);
      }
    };
    const _neutralizeSubtree = function _neutralizeSubtree2(root) {
      _walkTree(root, (node) => {
        if (_readNodeType(node) === NODE_TYPE.element) _stripAttributes(node, true);
      });
    };
    const _collectLiveNodes = function _collectLiveNodes2(root) {
      const nodes = [];
      _walkTree(root, (node) => {
        nodes.push(node);
      }, true);
      return nodes;
    };
    const _neutralizeEscapedElements = function _neutralizeEscapedElements2(root, before) {
      const reachable = setCreate();
      arrayForEach(_collectLiveNodes(root), (node) => {
        if (_readNodeType(node) === NODE_TYPE.element) setAdd(reachable, node);
      });
      arrayForEach(before, (node) => {
        if (_readNodeType(node) === NODE_TYPE.element && !setHas(reachable, node)) _stripAttributes(node, true);
      });
    };
    const _neutralizeDiscarded = function _neutralizeDiscarded2(root, before) {
      let rootRemoved = false;
      arrayForEach(DOMPurify.removed, (entry) => {
        if (entry.element) {
          rootRemoved = rootRemoved || entry.element === root;
          _neutralizeSubtree(entry.element);
        }
      });
      _neutralizeEscapedElements(root, before);
      return rootRemoved;
    };
    const _failClosed = function _failClosed2(root, before) {
      _neutralizeRoot(root);
      _neutralizeDiscarded(root, before);
    };
    const MAX_UNVISITED_ROUNDS = 8;
    const _sanitizeUnvisitedNodes = function _sanitizeUnvisitedNodes2(root, before) {
      const visited = IN_PLACE_VISITED;
      let original = null;
      const getOriginal = () => {
        if (!original) {
          original = setCreate();
          arrayForEach(before, (beforeNode) => {
            setAdd(original, beforeNode);
          });
        }
        return original;
      };
      for (let round = 0; ; round++) {
        const live = _collectLiveNodes(root);
        const containers = [root];
        let pending = false;
        arrayForEach(live, (node) => {
          if (_readNodeType(node) === NODE_TYPE.documentFragment) arrayPush(containers, node);
          else if (!pending && !setHas(visited, node)) pending = setHas(getOriginal(), node);
        });
        if (!pending) return;
        if (round >= MAX_UNVISITED_ROUNDS) throw typeErrorCreate("the tree kept changing during in-place sanitization; refusing to return it");
        arrayForEach(containers, (container, index) => {
          _sanitizeUnvisitedIn(container, index > 0, root, getOriginal());
        });
      }
    };
    const _sanitizeUnvisitedIn = function _sanitizeUnvisitedIn2(container, inShadow, root, original) {
      const iterator = _createNodeIterator(container);
      let node;
      while (node = iterator.nextNode()) {
        if (!setHas(original, node) || setHas(IN_PLACE_VISITED, node)) continue;
        if (inShadow) _executeHooks(hooks.uponSanitizeShadowNode, node, null);
        _sanitizeNode(node, root);
      }
    };
    const _markVisited = function _markVisited2(node) {
      if (!IN_PLACE_VISITED) return;
      setAdd(IN_PLACE_VISITED, node);
      const returns = mapGet(IN_PLACE_RETURNS, node);
      if (returns === void 0) return;
      if (returns >= MAX_UNVISITED_ROUNDS) throw typeErrorCreate("the tree kept changing during in-place sanitization; refusing to return it");
      mapSet(IN_PLACE_RETURNS, node, returns + 1);
    };
    const _unmarkVisited = function _unmarkVisited2(node) {
      if (!IN_PLACE_VISITED) return;
      setDelete(IN_PLACE_VISITED, node);
      if (mapGet(IN_PLACE_RETURNS, node) === void 0) mapSet(IN_PLACE_RETURNS, node, 0);
    };
    const _isPatchLinkageAttribute = function _isPatchLinkageAttribute2(lcName, lcTag) {
      if (!SAFE_FOR_XML) return false;
      if (lcName === "patchsrc") return true;
      return lcName === "for" && lcTag !== "label" && lcTag !== "output";
    };
    const _neutralizePatchLinkage = function _neutralizePatchLinkage2(root) {
      if (!SAFE_FOR_XML) return;
      _walkTree(root, (node) => {
        const nodeType = _readNodeType(node);
        if (nodeType === NODE_TYPE.processingInstruction || nodeType === NODE_TYPE.comment && regExpTest(COMMENT_MARKUP_PROBE, node.data)) {
          try {
            remove(node);
          } catch (_) {
          }
          return false;
        }
        if (nodeType === NODE_TYPE.element) {
          const element = node;
          const lcTag = transformCaseFunc(_readNodeName(node));
          try {
            if (element.hasAttribute && element.hasAttribute("patchsrc")) element.removeAttribute("patchsrc");
            if (element.hasAttribute && element.hasAttribute("for") && _isPatchLinkageAttribute("for", lcTag)) element.removeAttribute("for");
          } catch (_) {
          }
        }
        return true;
      });
    };
    const _initDocument = function _initDocument2(dirty) {
      let doc = null;
      let leadingWhitespace = null;
      if (FORCE_BODY) dirty = "<remove></remove>" + dirty;
      else {
        const matches = stringMatch(dirty, LEADING_WHITESPACE);
        leadingWhitespace = matches && matches[0];
      }
      if (PARSER_MEDIA_TYPE === "application/xhtml+xml" && NAMESPACE === HTML_NAMESPACE) dirty = '<html xmlns="http://www.w3.org/1999/xhtml"><head></head><body>' + dirty + "</body></html>";
      const dirtyPayload = trustedTypesPolicy ? _createTrustedHTML(dirty) : dirty;
      if (NAMESPACE === HTML_NAMESPACE) try {
        doc = new DOMParser().parseFromString(dirtyPayload, PARSER_MEDIA_TYPE);
      } catch (_) {
      }
      if (!doc || !doc.documentElement) {
        doc = implementation.createDocument(NAMESPACE, "template", null);
        try {
          doc.documentElement.innerHTML = IS_EMPTY_INPUT ? emptyHTML : dirtyPayload;
        } catch (_) {
        }
      }
      const body = doc.body || doc.documentElement;
      if (dirty && leadingWhitespace) body.insertBefore(document2.createTextNode(leadingWhitespace), body.childNodes[0] || null);
      if (NAMESPACE === HTML_NAMESPACE) return getElementsByTagName.call(doc, WHOLE_DOCUMENT ? "html" : "body")[0];
      return WHOLE_DOCUMENT ? doc.documentElement : body;
    };
    const _createNodeIterator = function _createNodeIterator2(root) {
      let characterDataOnly = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : false;
      const doc = getOwnerDocument ? getOwnerDocument(root) : root.ownerDocument;
      return createNodeIterator.call(doc || root, root, (characterDataOnly ? 0 : NodeFilter.SHOW_ELEMENT) | NodeFilter.SHOW_COMMENT | NodeFilter.SHOW_TEXT | NodeFilter.SHOW_PROCESSING_INSTRUCTION | NodeFilter.SHOW_CDATA_SECTION, null);
    };
    const _stripTemplateExpressions = function _stripTemplateExpressions2(value) {
      value = stringReplace(value, MUSTACHE_EXPR$1, " ");
      value = stringReplace(value, ERB_EXPR$1, " ");
      value = stringReplace(value, TMPLIT_EXPR$1, " ");
      return value;
    };
    const _scrubTemplateExpressions2 = function _scrubTemplateExpressions(node) {
      var _node$querySelectorAl;
      _scrubOwnCharacterData(node);
      const templates = (_node$querySelectorAl = node.querySelectorAll) === null || _node$querySelectorAl === void 0 ? void 0 : _node$querySelectorAl.call(node, "template");
      if (templates) arrayForEach(templates, (tmpl) => {
        if (_isDocumentFragment(tmpl.content)) _scrubTemplateExpressions2(tmpl.content);
      });
    };
    const TABLE_STRUCTURE = "table,tbody,thead,tfoot,tr,colgroup";
    const TABLE_CONTENT = addToSet({}, [
      "caption",
      "colgroup",
      "col",
      "thead",
      "tbody",
      "tfoot",
      "tr",
      "td",
      "th",
      "style",
      "script",
      "template"
    ]);
    const _fosterTableContent = function _fosterTableContent2(root) {
      var _root$querySelectorAl;
      if (PARSER_MEDIA_TYPE === "application/xhtml+xml") return;
      const structures = (_root$querySelectorAl = root.querySelectorAll) === null || _root$querySelectorAl === void 0 ? void 0 : _root$querySelectorAl.call(root, TABLE_STRUCTURE);
      if (!structures || structures.length === 0) return;
      arrayForEach(structures, (structure) => {
        arrayForEach(arraySlice(getChildNodes(structure)), (child) => {
          const nodeType = _readNodeType(child);
          if (!(nodeType === NODE_TYPE.text ? regExpTest(NON_WHITESPACE, child.data) : nodeType === NODE_TYPE.element && !TABLE_CONTENT[stringToLowerCase(_readNodeName(child))])) return;
          let table = structure;
          while (table && stringToLowerCase(_readNodeName(table)) !== "table") table = table === root ? null : getParentNode(table);
          const tableParent = table && table !== root ? getParentNode(table) : null;
          if (tableParent) try {
            tableParent.insertBefore(child, table);
          } catch (_) {
          }
        });
      });
    };
    const _scrubOwnCharacterData = function _scrubOwnCharacterData2(node) {
      _fosterTableContent(node);
      node.normalize();
      const walker = _createNodeIterator(node, true);
      let previousText = null;
      let currentNode = walker.nextNode();
      while (currentNode) {
        let data = currentNode.data;
        if (regExpTest(NUL, data)) {
          data = stringReplace(data, NUL, "");
          currentNode.data = data;
        }
        if (regExpTest(TMPL_MARKER, data)) {
          data = _stripTemplateExpressions(data);
          currentNode.data = data;
        }
        if (data && _readNodeType(currentNode) === NODE_TYPE.text) {
          if (previousText) {
            const tail = previousText.data;
            const boundary = tail[tail.length - 1] + data[0];
            if (regExpTest(TMPL_MARKER, boundary)) previousText.data = _stripTemplateExpressions(tail + data[0]);
          }
          previousText = currentNode;
        }
        currentNode = walker.nextNode();
      }
    };
    const _isClobbered = function _isClobbered2(element) {
      const realTagName = getNodeName ? getNodeName(element) : null;
      if (typeof realTagName !== "string") return false;
      if (transformCaseFunc(realTagName) !== "form") return false;
      return typeof element.nodeName !== "string" || typeof element.textContent !== "string" || typeof element.removeChild !== "function" || element.attributes !== getAttributes(element) || typeof element.removeAttribute !== "function" || typeof element.removeAttributeNode !== "function" || typeof element.getAttributeNode !== "function" || typeof element.setAttribute !== "function" || typeof element.namespaceURI !== "string" || typeof element.insertBefore !== "function" || typeof element.hasChildNodes !== "function" || element.nodeType !== getNodeType(element) || element.childNodes !== getChildNodes(element);
    };
    const _nodeTypeOf = function _nodeTypeOf2(value) {
      if (!getNodeType || typeof value !== "object" || value === null) return null;
      try {
        const nodeType = getNodeType(value);
        return typeof nodeType === "number" ? nodeType : null;
      } catch (_) {
        return null;
      }
    };
    const _isDocumentFragment = function _isDocumentFragment2(value) {
      return _nodeTypeOf(value) === NODE_TYPE.documentFragment;
    };
    const _isNode = function _isNode2(value) {
      return _nodeTypeOf(value) !== null;
    };
    function _executeHooks(hooks2, currentNode, data) {
      if (hooks2.length === 0) return;
      arrayForEach(hooks2, (hook) => {
        hook.call(DOMPurify, currentNode, data, CONFIG);
      });
    }
    const _isMalformedElement = function _isMalformedElement2(currentNode, tagName) {
      const isVoid = Boolean(VOID_ELEMENTS[tagName]);
      const literal = Boolean(LITERAL_TEXT_ELEMENTS[tagName]);
      if (!isVoid && !literal && !ESCAPABLE_TEXT_ELEMENTS[tagName]) return false;
      if (_readNamespaceURI(currentNode) !== HTML_NAMESPACE) return false;
      if (isVoid) return currentNode.hasChildNodes();
      if (_isNode(currentNode.firstElementChild)) return true;
      return literal && typeof currentNode.textContent === "string" && regExpTest(ELEMENT_MARKUP_PROBE, currentNode.textContent);
    };
    const _isUnsafeNode = function _isUnsafeNode2(currentNode, tagName, nodeType) {
      if (SAFE_FOR_XML && nodeType === NODE_TYPE.element && currentNode.hasChildNodes() && !_isNode(currentNode.firstElementChild) && regExpTest(ELEMENT_MARKUP_PROBE, currentNode.textContent) && regExpTest(ELEMENT_MARKUP_PROBE, currentNode.innerHTML)) return true;
      if (SAFE_FOR_XML && nodeType === NODE_TYPE.element && _isMalformedElement(currentNode, tagName)) return true;
      if (nodeType === NODE_TYPE.processingInstruction) return true;
      if (SAFE_FOR_XML && nodeType === NODE_TYPE.comment && regExpTest(COMMENT_UNSAFE, currentNode.data)) return true;
      return false;
    };
    const _matchesNameCheck = function _matchesNameCheck2(check, name) {
      if (check instanceof RegExp) return regExpTest(check, name);
      if (check instanceof Function) {
        for (var _len = arguments.length, args = new Array(_len > 2 ? _len - 2 : 0), _key = 2; _key < _len; _key++) args[_key - 2] = arguments[_key];
        return Boolean(check(name, ...args));
      }
      return false;
    };
    const _sanitizeDisallowedNode = function _sanitizeDisallowedNode2(currentNode, tagName, root) {
      if (!FORBID_TAGS[tagName] && _isBasicCustomElement(tagName) && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.tagNameCheck, tagName)) return false;
      if (KEEP_CONTENT && !FORBID_CONTENTS[tagName] && !FORBID_CONTENTS[stringToLowerCase(tagName)]) {
        const parentNode = getParentNode(currentNode);
        const childNodes = getChildNodes(currentNode);
        if (childNodes && parentNode) {
          const childCount = childNodes.length;
          for (let i = childCount - 1; i >= 0; --i) {
            const hoisted = currentNode === root ? cloneNode(childNodes[i], true) : childNodes[i];
            parentNode.insertBefore(hoisted, getNextSibling(currentNode));
          }
        }
      }
      _forceRemove(currentNode);
      return true;
    };
    const _forkSharedAllowlist = function _forkSharedAllowlist2(hookList, set2, defaultSet, setConfigSet) {
      if (hookList.length === 0) return set2;
      return set2 === defaultSet || set2 === setConfigSet ? cloneSet(set2) : set2;
    };
    const _handleHookDetachedNode = function _handleHookDetachedNode2(currentNode, root) {
      if (currentNode === root || getParentNode(currentNode) !== null) return false;
      if (IN_PLACE) _neutralizeSubtree(currentNode);
      _unmarkVisited(currentNode);
      return true;
    };
    const _sanitizeElements = function _sanitizeElements2(currentNode, root) {
      _markVisited(currentNode);
      _executeHooks(hooks.beforeSanitizeElements, currentNode, null);
      if (_handleHookDetachedNode(currentNode, root)) return true;
      if (_isClobbered(currentNode)) {
        _forceRemove(currentNode);
        return true;
      }
      const tagName = transformCaseFunc(_readNodeName(currentNode));
      const nodeType = _readNodeType(currentNode);
      ALLOWED_TAGS = _forkSharedAllowlist(hooks.uponSanitizeElement, ALLOWED_TAGS, DEFAULT_ALLOWED_TAGS, SET_CONFIG_ALLOWED_TAGS);
      _executeHooks(hooks.uponSanitizeElement, currentNode, {
        tagName,
        allowedTags: ALLOWED_TAGS
      });
      if (_handleHookDetachedNode(currentNode, root)) return true;
      if (_isUnsafeNode(currentNode, tagName, nodeType)) {
        _forceRemove(currentNode);
        return true;
      }
      if ((FORBID_TAGS[tagName] || !(EXTRA_ELEMENT_HANDLING.tagCheck instanceof Function && EXTRA_ELEMENT_HANDLING.tagCheck(tagName)) && !ALLOWED_TAGS[tagName]) && _sanitizeDisallowedNode(currentNode, tagName, root)) return true;
      if (nodeType === NODE_TYPE.element && !_checkValidNamespace(currentNode)) {
        _forceRemove(currentNode);
        return true;
      }
      if ((tagName === "noscript" || tagName === "noembed" || tagName === "noframes") && regExpTest(FALLBACK_TAG_CLOSE, currentNode.innerHTML)) {
        _forceRemove(currentNode);
        return true;
      }
      if (SAFE_FOR_TEMPLATES && nodeType === NODE_TYPE.text) {
        const text2 = currentNode.textContent;
        const content = _stripTemplateExpressions(text2);
        if (text2 !== content) {
          arrayPush(DOMPurify.removed, { element: currentNode.cloneNode() });
          currentNode.textContent = content;
        }
      }
      _executeHooks(hooks.afterSanitizeElements, currentNode, null);
      return _handleHookDetachedNode(currentNode, root);
    };
    const _isValidAttribute = function _isValidAttribute2(lcTag, lcName, value) {
      if (FORBID_ATTR[lcName]) return false;
      if (_isPatchLinkageAttribute(lcName, lcTag)) return false;
      if (SANITIZE_DOM && (lcName === "id" || lcName === "name") && (value in document2 || value in formElement)) return false;
      const nameIsPermitted = ALLOWED_ATTR[lcName] || EXTRA_ELEMENT_HANDLING.attributeCheck instanceof Function && EXTRA_ELEMENT_HANDLING.attributeCheck(lcName, lcTag);
      if (ALLOW_DATA_ATTR && regExpTest(DATA_ATTR$1, lcName)) return true;
      if (ALLOW_ARIA_ATTR && regExpTest(ARIA_ATTR$1, lcName)) return true;
      if (!nameIsPermitted) return _isBasicCustomElement(lcTag) && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.tagNameCheck, lcTag) && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.attributeNameCheck, lcName, lcTag) || lcName === "is" && CUSTOM_ELEMENT_HANDLING.allowCustomizedBuiltInElements && _matchesNameCheck(CUSTOM_ELEMENT_HANDLING.tagNameCheck, value);
      if (URI_SAFE_ATTRIBUTES[lcName]) return true;
      const compactValue = stringReplace(value, ATTR_WHITESPACE$1, "");
      if (regExpTest(IS_ALLOWED_URI$1, compactValue)) return true;
      if ((lcName === "src" || lcName === "xlink:href" || lcName === "href") && lcTag !== "script" && stringIndexOf(value, "data:") === 0 && DATA_URI_TAGS[lcTag]) return true;
      if (ALLOW_UNKNOWN_PROTOCOLS && !regExpTest(IS_SCRIPT_OR_DATA$1, compactValue)) return true;
      return !value;
    };
    const RESERVED_CUSTOM_ELEMENT_NAMES = addToSet({}, [
      "annotation-xml",
      "color-profile",
      "font-face",
      "font-face-format",
      "font-face-name",
      "font-face-src",
      "font-face-uri",
      "missing-glyph"
    ]);
    const _isBasicCustomElement = function _isBasicCustomElement2(tagName) {
      return !RESERVED_CUSTOM_ELEMENT_NAMES[stringToLowerCase(tagName)] && regExpTest(CUSTOM_ELEMENT$1, tagName);
    };
    const _applyTrustedTypesToAttribute = function _applyTrustedTypesToAttribute2(lcTag, lcName, namespaceURI, value) {
      if (trustedTypesPolicy && typeof trustedTypes === "object" && typeof trustedTypes.getAttributeType === "function" && !namespaceURI) switch (trustedTypes.getAttributeType(lcTag, lcName)) {
        case "TrustedHTML":
          return _createTrustedHTML(value);
        case "TrustedScriptURL":
          return _createTrustedScriptURL(value);
      }
      return value;
    };
    const _setAttributeValue = function _setAttributeValue2(currentNode, name, namespaceURI, value) {
      try {
        if (namespaceURI) currentNode.setAttributeNS(namespaceURI, name, value);
        else currentNode.setAttribute(name, value);
        if (_isClobbered(currentNode)) {
          _forceRemove(currentNode);
          return false;
        }
        return true;
      } catch (_) {
        _removeAttribute(name, currentNode);
        return false;
      }
    };
    const _isUnsafeAttributeValue = function _isUnsafeAttributeValue2(name, value) {
      if (SAFE_FOR_XML && regExpTest(ATTR_VALUE_BREAKOUT, value)) return true;
      return stringToLowerCase(name) === "attributename" && stringIndexOf(value, "href") !== -1;
    };
    const _sanitizeAttributes = function _sanitizeAttributes2(currentNode, root) {
      _executeHooks(hooks.beforeSanitizeAttributes, currentNode, null);
      if (_handleHookDetachedNode(currentNode, root)) return;
      const attributes = _readAttributes(currentNode);
      if (!attributes || _isClobbered(currentNode)) return;
      ALLOWED_ATTR = _forkSharedAllowlist(hooks.uponSanitizeAttribute, ALLOWED_ATTR, DEFAULT_ALLOWED_ATTR, SET_CONFIG_ALLOWED_ATTR);
      const hookEvent = {
        attrName: "",
        attrValue: "",
        keepAttr: true,
        allowedAttributes: ALLOWED_ATTR,
        forceKeepAttr: void 0
      };
      let l = attributes.length;
      const lcTag = transformCaseFunc(_readNodeName(currentNode));
      while (l--) {
        const attr = attributes[l];
        if (!attr) continue;
        const name = attr.name, namespaceURI = attr.namespaceURI, attrValue = attr.value;
        const lcName = transformCaseFunc(name);
        const initValue = attrValue;
        let value = name === "value" ? initValue : stringTrim(initValue);
        let recreatedNamedProp = false;
        hookEvent.attrName = lcName;
        hookEvent.attrValue = value;
        hookEvent.keepAttr = true;
        hookEvent.forceKeepAttr = void 0;
        _executeHooks(hooks.uponSanitizeAttribute, currentNode, hookEvent);
        value = hookEvent.attrValue;
        if (SANITIZE_NAMED_PROPS && (lcName === "id" || lcName === "name") && stringIndexOf(value, SANITIZE_NAMED_PROPS_PREFIX) !== 0) {
          _removeAttribute(name, currentNode, attr);
          value = SANITIZE_NAMED_PROPS_PREFIX + value;
          recreatedNamedProp = true;
        }
        if (_isUnsafeAttributeValue(name, value)) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        if (hookEvent.forceKeepAttr) continue;
        if (!hookEvent.keepAttr) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        if (!ALLOW_SELF_CLOSE_IN_ATTR && regExpTest(SELF_CLOSING_TAG, value)) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        if (SAFE_FOR_TEMPLATES) value = _stripTemplateExpressions(value);
        if (!_isValidAttribute(lcTag, lcName, value)) {
          _removeAttribute(name, currentNode, attr);
          continue;
        }
        value = _applyTrustedTypesToAttribute(lcTag, lcName, namespaceURI, value);
        if (value !== initValue) {
          if (_setAttributeValue(currentNode, name, namespaceURI, value) && recreatedNamedProp) arrayPop(DOMPurify.removed);
        }
      }
      _executeHooks(hooks.afterSanitizeAttributes, currentNode, null);
      _handleHookDetachedNode(currentNode, root);
    };
    const _sanitizeNode = function _sanitizeNode2(node, root) {
      _sanitizeElements(node, root);
      _sanitizeAttributes(node, root);
      if (_isDocumentFragment(node.content)) _sanitizeShadowDOM2(node.content);
    };
    const _sanitizeShadowDOM2 = function _sanitizeShadowDOM(fragment) {
      let shadowNode = null;
      const shadowIterator = _createNodeIterator(fragment);
      _executeHooks(hooks.beforeSanitizeShadowDOM, fragment, null);
      while (shadowNode = shadowIterator.nextNode()) {
        _executeHooks(hooks.uponSanitizeShadowNode, shadowNode, null);
        _sanitizeNode(shadowNode, fragment);
        if (_readNodeType(shadowNode) === NODE_TYPE.element) {
          const innerSr = getShadowRoot(shadowNode);
          if (_isDocumentFragment(innerSr)) {
            _sanitizeAttachedShadowRoots(innerSr);
            _sanitizeShadowDOM2(innerSr);
          }
        }
      }
      if (SAFE_FOR_TEMPLATES) _scrubOwnCharacterData(fragment);
      _executeHooks(hooks.afterSanitizeShadowDOM, fragment, null);
    };
    const _sanitizeAttachedShadowRoots = function _sanitizeAttachedShadowRoots2(root) {
      const stack = [{
        node: root,
        shadow: null
      }];
      while (stack.length > 0) {
        const item = stack.pop();
        if (item.shadow) {
          _sanitizeShadowDOM2(item.shadow);
          continue;
        }
        const node = item.node;
        const isElement = _readNodeType(node) === NODE_TYPE.element;
        const childNodes = getChildNodes(node);
        if (childNodes) for (let i = childNodes.length - 1; i >= 0; --i) stack.push({
          node: childNodes[i],
          shadow: null
        });
        if (isElement) {
          const rootName = getNodeName ? getNodeName(node) : null;
          if (typeof rootName === "string" && transformCaseFunc(rootName) === "template") {
            const content = node.content;
            if (_isDocumentFragment(content)) stack.push({
              node: content,
              shadow: null
            });
          }
        }
        if (isElement) {
          const sr = getShadowRoot(node);
          if (_isDocumentFragment(sr)) stack.push({
            node: null,
            shadow: sr
          }, {
            node: sr,
            shadow: null
          });
        }
      }
    };
    const _sanitize = function _sanitize2(dirty) {
      let cfg = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
      let body = null;
      let importedNode = null;
      let currentNode = null;
      let returnNode = null;
      IS_EMPTY_INPUT = !dirty;
      if (IS_EMPTY_INPUT) dirty = "<!-->";
      if (typeof dirty !== "string" && !_isNode(dirty)) {
        dirty = stringifyValue(dirty);
        if (typeof dirty !== "string") throw typeErrorCreate("dirty is not a string, aborting");
      }
      if (!DOMPurify.isSupported) return dirty;
      if (SET_CONFIG) {
        ALLOWED_TAGS = SET_CONFIG_ALLOWED_TAGS;
        ALLOWED_ATTR = SET_CONFIG_ALLOWED_ATTR;
      } else _parseConfig(cfg);
      ACTIVE_CONFIG = CONFIG;
      DOMPurify.removed = [];
      const inPlace = IN_PLACE && typeof dirty !== "string" && _isNode(dirty);
      let inPlaceNodes = [];
      const outerVisited = IN_PLACE_VISITED;
      const outerReturns = IN_PLACE_RETURNS;
      if (inPlace) {
        _neutralizePatchLinkage(dirty);
        inPlaceNodes = _collectLiveNodes(dirty);
        const nn = _readNodeName(dirty);
        if (typeof nn === "string") {
          const tagName = transformCaseFunc(nn);
          if (!ALLOWED_TAGS[tagName] || FORBID_TAGS[tagName]) {
            _failClosed(dirty, inPlaceNodes);
            throw typeErrorCreate("root node is forbidden and cannot be sanitized in-place");
          }
        }
        if (_isClobbered(dirty)) {
          _failClosed(dirty, inPlaceNodes);
          throw typeErrorCreate("root node is clobbered and cannot be sanitized in-place");
        }
        IN_PLACE_VISITED = setCreate();
        IN_PLACE_RETURNS = mapCreate();
      } else if (_isNode(dirty)) {
        body = _initDocument("<!---->");
        importedNode = body.ownerDocument.importNode(dirty, true);
        if (importedNode.nodeType === NODE_TYPE.element && importedNode.nodeName === "BODY") body = importedNode;
        else if (importedNode.nodeName === "HTML") body = importedNode;
        else body.appendChild(importedNode);
        _sanitizeAttachedShadowRoots(body);
      } else {
        if (!RETURN_DOM && !SAFE_FOR_TEMPLATES && !WHOLE_DOCUMENT && dirty.indexOf("<") === -1) return _signOutput(dirty);
        body = _initDocument(dirty);
        if (!body) return RETURN_DOM ? null : RETURN_TRUSTED_TYPE ? emptyHTML : "";
      }
      if (body && FORCE_BODY) _forceRemove(body.firstChild);
      const walkRoot = inPlace ? dirty : body;
      try {
        if (inPlace) _sanitizeAttachedShadowRoots(dirty);
        const nodeIterator = _createNodeIterator(walkRoot);
        while (currentNode = nodeIterator.nextNode()) _sanitizeNode(currentNode, walkRoot);
        if (inPlace) _sanitizeUnvisitedNodes(dirty, inPlaceNodes);
      } catch (error) {
        if (inPlace) _failClosed(dirty, inPlaceNodes);
        throw error;
      } finally {
        IN_PLACE_VISITED = outerVisited;
        IN_PLACE_RETURNS = outerReturns;
      }
      if (inPlace) {
        if (_neutralizeDiscarded(dirty, inPlaceNodes)) throw typeErrorCreate("a node selected for removal could not be safely returned; refusing to sanitize in place");
        if (SAFE_FOR_TEMPLATES) _scrubTemplateExpressions2(dirty);
        return dirty;
      }
      if (RETURN_DOM) {
        if (SAFE_FOR_TEMPLATES) _scrubTemplateExpressions2(body);
        if (RETURN_DOM_FRAGMENT) {
          returnNode = createDocumentFragment.call(body.ownerDocument);
          while (body.firstChild) returnNode.appendChild(body.firstChild);
        } else returnNode = body;
        if (ALLOWED_ATTR.shadowroot || ALLOWED_ATTR.shadowrootmode) returnNode = importNode.call(originalDocument, returnNode, true);
        return returnNode;
      }
      if (SAFE_FOR_TEMPLATES) _scrubOwnCharacterData(body);
      let serializedHTML = WHOLE_DOCUMENT ? body.outerHTML : body.innerHTML;
      if (WHOLE_DOCUMENT && ALLOWED_TAGS["!doctype"] && body.ownerDocument && body.ownerDocument.doctype && body.ownerDocument.doctype.name && regExpTest(DOCTYPE_NAME, body.ownerDocument.doctype.name)) serializedHTML = "<!DOCTYPE " + body.ownerDocument.doctype.name + ">\n" + serializedHTML;
      return _signOutput(serializedHTML);
    };
    DOMPurify.sanitize = function(dirty) {
      let cfg = arguments.length > 1 && arguments[1] !== void 0 ? arguments[1] : {};
      const outerConfig = ACTIVE_CONFIG;
      const outerTags = ALLOWED_TAGS;
      const outerAttr = ALLOWED_ATTR;
      const outerRemoved = DOMPurify.removed;
      ACTIVE_CONFIG = null;
      try {
        return _sanitize(dirty, cfg);
      } finally {
        ACTIVE_CONFIG = outerConfig;
        if (outerConfig) {
          if (!SET_CONFIG) _parseConfig(outerConfig);
          ALLOWED_TAGS = outerTags;
          ALLOWED_ATTR = outerAttr;
          DOMPurify.removed = outerRemoved;
        }
      }
    };
    DOMPurify.setConfig = function() {
      let cfg = arguments.length > 0 && arguments[0] !== void 0 ? arguments[0] : {};
      _parseConfig(cfg);
      SET_CONFIG = true;
      SET_CONFIG_ALLOWED_TAGS = ALLOWED_TAGS;
      SET_CONFIG_ALLOWED_ATTR = ALLOWED_ATTR;
    };
    DOMPurify.clearConfig = function() {
      CONFIG = null;
      SET_CONFIG = false;
      SET_CONFIG_ALLOWED_TAGS = null;
      SET_CONFIG_ALLOWED_ATTR = null;
      trustedTypesPolicy = defaultTrustedTypesPolicy;
      emptyHTML = "";
    };
    DOMPurify.isValidAttribute = function(tag2, attr, value) {
      if (!CONFIG) _parseConfig({});
      const lcTag = transformCaseFunc(tag2);
      const lcName = transformCaseFunc(attr);
      if (typeof value === "string" && (_isUnsafeAttributeValue(attr, value) || !ALLOW_SELF_CLOSE_IN_ATTR && regExpTest(SELF_CLOSING_TAG, value))) return false;
      return _isValidAttribute(lcTag, lcName, value);
    };
    DOMPurify.addHook = function(entryPoint, hookFunction) {
      if (typeof hookFunction !== "function") return;
      if (!objectHasOwnProperty(hooks, entryPoint)) return;
      arrayPush(hooks[entryPoint], hookFunction);
    };
    DOMPurify.removeHook = function(entryPoint, hookFunction) {
      if (!objectHasOwnProperty(hooks, entryPoint)) return;
      if (hookFunction !== void 0) {
        const index = arrayLastIndexOf(hooks[entryPoint], hookFunction);
        return index === -1 ? void 0 : arraySplice(hooks[entryPoint], index, 1)[0];
      }
      return arrayPop(hooks[entryPoint]);
    };
    DOMPurify.removeHooks = function(entryPoint) {
      if (!objectHasOwnProperty(hooks, entryPoint)) return;
      hooks[entryPoint] = [];
    };
    DOMPurify.removeAllHooks = function() {
      hooks = _createHooksMap();
    };
    return DOMPurify;
  }
  var purify_default = createDOMPurify();

  // node_modules/js-yaml/dist/js-yaml.mjs
  function getDefaultExportFromCjs(x) {
    return x && x.__esModule && Object.prototype.hasOwnProperty.call(x, "default") ? x["default"] : x;
  }
  var jsYaml = {};
  var loader = {};
  var common = {};
  var hasRequiredCommon;
  function requireCommon() {
    if (hasRequiredCommon) return common;
    hasRequiredCommon = 1;
    function isNothing(subject) {
      return typeof subject === "undefined" || subject === null;
    }
    function isObject(subject) {
      return typeof subject === "object" && subject !== null;
    }
    function toArray(sequence) {
      if (Array.isArray(sequence)) return sequence;
      else if (isNothing(sequence)) return [];
      return [sequence];
    }
    function extend(target, source) {
      if (source) {
        const sourceKeys = Object.keys(source);
        for (let index = 0, length = sourceKeys.length; index < length; index += 1) {
          const key = sourceKeys[index];
          target[key] = source[key];
        }
      }
      return target;
    }
    function repeat(string, count) {
      let result = "";
      for (let cycle = 0; cycle < count; cycle += 1) {
        result += string;
      }
      return result;
    }
    function isNegativeZero(number) {
      return number === 0 && Number.NEGATIVE_INFINITY === 1 / number;
    }
    common.isNothing = isNothing;
    common.isObject = isObject;
    common.toArray = toArray;
    common.repeat = repeat;
    common.isNegativeZero = isNegativeZero;
    common.extend = extend;
    return common;
  }
  var exception;
  var hasRequiredException;
  function requireException() {
    if (hasRequiredException) return exception;
    hasRequiredException = 1;
    function formatError(exception2, compact) {
      let where = "";
      const message = exception2.reason || "(unknown reason)";
      if (!exception2.mark) return message;
      if (exception2.mark.name) {
        where += 'in "' + exception2.mark.name + '" ';
      }
      where += "(" + (exception2.mark.line + 1) + ":" + (exception2.mark.column + 1) + ")";
      if (!compact && exception2.mark.snippet) {
        where += "\n\n" + exception2.mark.snippet;
      }
      return message + " " + where;
    }
    function YAMLException2(reason, mark) {
      Error.call(this);
      this.name = "YAMLException";
      this.reason = reason;
      this.mark = mark;
      this.message = formatError(this, false);
      if (Error.captureStackTrace) {
        Error.captureStackTrace(this, this.constructor);
      } else {
        this.stack = new Error().stack || "";
      }
    }
    YAMLException2.prototype = Object.create(Error.prototype);
    YAMLException2.prototype.constructor = YAMLException2;
    YAMLException2.prototype.toString = function toString(compact) {
      return this.name + ": " + formatError(this, compact);
    };
    exception = YAMLException2;
    return exception;
  }
  var snippet;
  var hasRequiredSnippet;
  function requireSnippet() {
    if (hasRequiredSnippet) return snippet;
    hasRequiredSnippet = 1;
    const common2 = requireCommon();
    function getLine(buffer, lineStart, lineEnd, position, maxLineLength) {
      let head = "";
      let tail = "";
      const maxHalfLength = Math.floor(maxLineLength / 2) - 1;
      if (position - lineStart > maxHalfLength) {
        head = " ... ";
        lineStart = position - maxHalfLength + head.length;
      }
      if (lineEnd - position > maxHalfLength) {
        tail = " ...";
        lineEnd = position + maxHalfLength - tail.length;
      }
      return {
        str: head + buffer.slice(lineStart, lineEnd).replace(/\t/g, "\u2192") + tail,
        pos: position - lineStart + head.length
        // relative position
      };
    }
    function padStart(string, max) {
      return common2.repeat(" ", max - string.length) + string;
    }
    function makeSnippet(mark, options2) {
      options2 = Object.create(options2 || null);
      if (!mark.buffer) return null;
      if (!options2.maxLength) options2.maxLength = 79;
      if (typeof options2.indent !== "number") options2.indent = 1;
      if (typeof options2.linesBefore !== "number") options2.linesBefore = 3;
      if (typeof options2.linesAfter !== "number") options2.linesAfter = 2;
      const re = /\r?\n|\r|\0/g;
      const lineStarts = [0];
      const lineEnds = [];
      let match;
      let foundLineNo = -1;
      while (match = re.exec(mark.buffer)) {
        lineEnds.push(match.index);
        lineStarts.push(match.index + match[0].length);
        if (mark.position <= match.index && foundLineNo < 0) {
          foundLineNo = lineStarts.length - 2;
        }
      }
      if (foundLineNo < 0) foundLineNo = lineStarts.length - 1;
      let result = "";
      const lineNoLength = Math.min(mark.line + options2.linesAfter, lineEnds.length).toString().length;
      const maxLineLength = options2.maxLength - (options2.indent + lineNoLength + 3);
      for (let i = 1; i <= options2.linesBefore; i++) {
        if (foundLineNo - i < 0) break;
        const line2 = getLine(
          mark.buffer,
          lineStarts[foundLineNo - i],
          lineEnds[foundLineNo - i],
          mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo - i]),
          maxLineLength
        );
        result = common2.repeat(" ", options2.indent) + padStart((mark.line - i + 1).toString(), lineNoLength) + " | " + line2.str + "\n" + result;
      }
      const line = getLine(mark.buffer, lineStarts[foundLineNo], lineEnds[foundLineNo], mark.position, maxLineLength);
      result += common2.repeat(" ", options2.indent) + padStart((mark.line + 1).toString(), lineNoLength) + " | " + line.str + "\n";
      result += common2.repeat("-", options2.indent + lineNoLength + 3 + line.pos) + "^\n";
      for (let i = 1; i <= options2.linesAfter; i++) {
        if (foundLineNo + i >= lineEnds.length) break;
        const line2 = getLine(
          mark.buffer,
          lineStarts[foundLineNo + i],
          lineEnds[foundLineNo + i],
          mark.position - (lineStarts[foundLineNo] - lineStarts[foundLineNo + i]),
          maxLineLength
        );
        result += common2.repeat(" ", options2.indent) + padStart((mark.line + i + 1).toString(), lineNoLength) + " | " + line2.str + "\n";
      }
      return result.replace(/\n$/, "");
    }
    snippet = makeSnippet;
    return snippet;
  }
  var type;
  var hasRequiredType;
  function requireType() {
    if (hasRequiredType) return type;
    hasRequiredType = 1;
    const YAMLException2 = requireException();
    const TYPE_CONSTRUCTOR_OPTIONS = [
      "kind",
      "multi",
      "resolve",
      "construct",
      "instanceOf",
      "predicate",
      "represent",
      "representName",
      "defaultStyle",
      "styleAliases"
    ];
    const YAML_NODE_KINDS = [
      "scalar",
      "sequence",
      "mapping"
    ];
    function compileStyleAliases(map2) {
      const result = {};
      if (map2 !== null) {
        Object.keys(map2).forEach(function(style) {
          map2[style].forEach(function(alias) {
            result[String(alias)] = style;
          });
        });
      }
      return result;
    }
    function Type2(tag2, options2) {
      options2 = options2 || {};
      Object.keys(options2).forEach(function(name) {
        if (TYPE_CONSTRUCTOR_OPTIONS.indexOf(name) === -1) {
          throw new YAMLException2('Unknown option "' + name + '" is met in definition of "' + tag2 + '" YAML type.');
        }
      });
      this.options = options2;
      this.tag = tag2;
      this.kind = options2["kind"] || null;
      this.resolve = options2["resolve"] || function() {
        return true;
      };
      this.construct = options2["construct"] || function(data) {
        return data;
      };
      this.instanceOf = options2["instanceOf"] || null;
      this.predicate = options2["predicate"] || null;
      this.represent = options2["represent"] || null;
      this.representName = options2["representName"] || null;
      this.defaultStyle = options2["defaultStyle"] || null;
      this.multi = options2["multi"] || false;
      this.styleAliases = compileStyleAliases(options2["styleAliases"] || null);
      if (YAML_NODE_KINDS.indexOf(this.kind) === -1) {
        throw new YAMLException2('Unknown kind "' + this.kind + '" is specified for "' + tag2 + '" YAML type.');
      }
    }
    type = Type2;
    return type;
  }
  var schema;
  var hasRequiredSchema;
  function requireSchema() {
    if (hasRequiredSchema) return schema;
    hasRequiredSchema = 1;
    const YAMLException2 = requireException();
    const Type2 = requireType();
    function compileList(schema2, name) {
      const result = [];
      schema2[name].forEach(function(currentType) {
        let newIndex = result.length;
        result.forEach(function(previousType, previousIndex) {
          if (previousType.tag === currentType.tag && previousType.kind === currentType.kind && previousType.multi === currentType.multi) {
            newIndex = previousIndex;
          }
        });
        result[newIndex] = currentType;
      });
      return result;
    }
    function compileMap() {
      const result = {
        scalar: {},
        sequence: {},
        mapping: {},
        fallback: {},
        multi: {
          scalar: [],
          sequence: [],
          mapping: [],
          fallback: []
        }
      };
      function collectType(type2) {
        if (type2.multi) {
          result.multi[type2.kind].push(type2);
          result.multi["fallback"].push(type2);
        } else {
          result[type2.kind][type2.tag] = result["fallback"][type2.tag] = type2;
        }
      }
      for (let index = 0, length = arguments.length; index < length; index += 1) {
        arguments[index].forEach(collectType);
      }
      return result;
    }
    function Schema2(definition) {
      return this.extend(definition);
    }
    Schema2.prototype.extend = function extend(definition) {
      let implicit = [];
      let explicit = [];
      if (definition instanceof Type2) {
        explicit.push(definition);
      } else if (Array.isArray(definition)) {
        explicit = explicit.concat(definition);
      } else if (definition && (Array.isArray(definition.implicit) || Array.isArray(definition.explicit))) {
        if (definition.implicit) implicit = implicit.concat(definition.implicit);
        if (definition.explicit) explicit = explicit.concat(definition.explicit);
      } else {
        throw new YAMLException2("Schema.extend argument should be a Type, [ Type ], or a schema definition ({ implicit: [...], explicit: [...] })");
      }
      implicit.forEach(function(type2) {
        if (!(type2 instanceof Type2)) {
          throw new YAMLException2("Specified list of YAML types (or a single Type object) contains a non-Type object.");
        }
        if (type2.loadKind && type2.loadKind !== "scalar") {
          throw new YAMLException2("There is a non-scalar type in the implicit list of a schema. Implicit resolving of such types is not supported.");
        }
        if (type2.multi) {
          throw new YAMLException2("There is a multi type in the implicit list of a schema. Multi tags can only be listed as explicit.");
        }
      });
      explicit.forEach(function(type2) {
        if (!(type2 instanceof Type2)) {
          throw new YAMLException2("Specified list of YAML types (or a single Type object) contains a non-Type object.");
        }
      });
      const result = Object.create(Schema2.prototype);
      result.implicit = (this.implicit || []).concat(implicit);
      result.explicit = (this.explicit || []).concat(explicit);
      result.compiledImplicit = compileList(result, "implicit");
      result.compiledExplicit = compileList(result, "explicit");
      result.compiledTypeMap = compileMap(result.compiledImplicit, result.compiledExplicit);
      return result;
    };
    schema = Schema2;
    return schema;
  }
  var str;
  var hasRequiredStr;
  function requireStr() {
    if (hasRequiredStr) return str;
    hasRequiredStr = 1;
    const Type2 = requireType();
    str = new Type2("tag:yaml.org,2002:str", {
      kind: "scalar",
      construct: function(data) {
        return data !== null ? data : "";
      }
    });
    return str;
  }
  var seq;
  var hasRequiredSeq;
  function requireSeq() {
    if (hasRequiredSeq) return seq;
    hasRequiredSeq = 1;
    const Type2 = requireType();
    seq = new Type2("tag:yaml.org,2002:seq", {
      kind: "sequence",
      construct: function(data) {
        return data !== null ? data : [];
      }
    });
    return seq;
  }
  var map;
  var hasRequiredMap;
  function requireMap() {
    if (hasRequiredMap) return map;
    hasRequiredMap = 1;
    const Type2 = requireType();
    map = new Type2("tag:yaml.org,2002:map", {
      kind: "mapping",
      construct: function(data) {
        return data !== null ? data : {};
      }
    });
    return map;
  }
  var failsafe;
  var hasRequiredFailsafe;
  function requireFailsafe() {
    if (hasRequiredFailsafe) return failsafe;
    hasRequiredFailsafe = 1;
    const Schema2 = requireSchema();
    failsafe = new Schema2({
      explicit: [
        requireStr(),
        requireSeq(),
        requireMap()
      ]
    });
    return failsafe;
  }
  var _null;
  var hasRequired_null;
  function require_null() {
    if (hasRequired_null) return _null;
    hasRequired_null = 1;
    const Type2 = requireType();
    function resolveYamlNull(data) {
      if (data === null) return true;
      const max = data.length;
      return max === 1 && data === "~" || max === 4 && (data === "null" || data === "Null" || data === "NULL");
    }
    function constructYamlNull() {
      return null;
    }
    function isNull(object) {
      return object === null;
    }
    _null = new Type2("tag:yaml.org,2002:null", {
      kind: "scalar",
      resolve: resolveYamlNull,
      construct: constructYamlNull,
      predicate: isNull,
      represent: {
        canonical: function() {
          return "~";
        },
        lowercase: function() {
          return "null";
        },
        uppercase: function() {
          return "NULL";
        },
        camelcase: function() {
          return "Null";
        },
        empty: function() {
          return "";
        }
      },
      defaultStyle: "lowercase"
    });
    return _null;
  }
  var bool;
  var hasRequiredBool;
  function requireBool() {
    if (hasRequiredBool) return bool;
    hasRequiredBool = 1;
    const Type2 = requireType();
    function resolveYamlBoolean(data) {
      if (data === null) return false;
      const max = data.length;
      return max === 4 && (data === "true" || data === "True" || data === "TRUE") || max === 5 && (data === "false" || data === "False" || data === "FALSE");
    }
    function constructYamlBoolean(data) {
      return data === "true" || data === "True" || data === "TRUE";
    }
    function isBoolean(object) {
      return Object.prototype.toString.call(object) === "[object Boolean]";
    }
    bool = new Type2("tag:yaml.org,2002:bool", {
      kind: "scalar",
      resolve: resolveYamlBoolean,
      construct: constructYamlBoolean,
      predicate: isBoolean,
      represent: {
        lowercase: function(object) {
          return object ? "true" : "false";
        },
        uppercase: function(object) {
          return object ? "TRUE" : "FALSE";
        },
        camelcase: function(object) {
          return object ? "True" : "False";
        }
      },
      defaultStyle: "lowercase"
    });
    return bool;
  }
  var int;
  var hasRequiredInt;
  function requireInt() {
    if (hasRequiredInt) return int;
    hasRequiredInt = 1;
    const common2 = requireCommon();
    const Type2 = requireType();
    function isHexCode(c) {
      return c >= 48 && c <= 57 || c >= 65 && c <= 70 || c >= 97 && c <= 102;
    }
    function isOctCode(c) {
      return c >= 48 && c <= 55;
    }
    function isDecCode(c) {
      return c >= 48 && c <= 57;
    }
    function resolveYamlInteger(data) {
      if (data === null) return false;
      const max = data.length;
      let index = 0;
      let hasDigits = false;
      if (!max) return false;
      let ch = data[index];
      if (ch === "-" || ch === "+") {
        ch = data[++index];
      }
      if (ch === "0") {
        if (index + 1 === max) return true;
        ch = data[++index];
        if (ch === "b") {
          index++;
          for (; index < max; index++) {
            ch = data[index];
            if (ch !== "0" && ch !== "1") return false;
            hasDigits = true;
          }
          return hasDigits && isFinite(parseYamlInteger(data));
        }
        if (ch === "x") {
          index++;
          for (; index < max; index++) {
            if (!isHexCode(data.charCodeAt(index))) return false;
            hasDigits = true;
          }
          return hasDigits && isFinite(parseYamlInteger(data));
        }
        if (ch === "o") {
          index++;
          for (; index < max; index++) {
            if (!isOctCode(data.charCodeAt(index))) return false;
            hasDigits = true;
          }
          return hasDigits && isFinite(parseYamlInteger(data));
        }
      }
      for (; index < max; index++) {
        if (!isDecCode(data.charCodeAt(index))) {
          return false;
        }
        hasDigits = true;
      }
      if (!hasDigits) return false;
      return isFinite(parseYamlInteger(data));
    }
    function parseYamlInteger(data) {
      let value = data;
      let sign = 1;
      let ch = value[0];
      if (ch === "-" || ch === "+") {
        if (ch === "-") sign = -1;
        value = value.slice(1);
        ch = value[0];
      }
      if (value === "0") return 0;
      if (ch === "0") {
        if (value[1] === "b") return sign * parseInt(value.slice(2), 2);
        if (value[1] === "x") return sign * parseInt(value.slice(2), 16);
        if (value[1] === "o") return sign * parseInt(value.slice(2), 8);
      }
      return sign * parseInt(value, 10);
    }
    function constructYamlInteger(data) {
      return parseYamlInteger(data);
    }
    function isInteger2(object) {
      return Object.prototype.toString.call(object) === "[object Number]" && (object % 1 === 0 && !common2.isNegativeZero(object));
    }
    int = new Type2("tag:yaml.org,2002:int", {
      kind: "scalar",
      resolve: resolveYamlInteger,
      construct: constructYamlInteger,
      predicate: isInteger2,
      represent: {
        binary: function(obj) {
          return obj >= 0 ? "0b" + obj.toString(2) : "-0b" + obj.toString(2).slice(1);
        },
        octal: function(obj) {
          return obj >= 0 ? "0o" + obj.toString(8) : "-0o" + obj.toString(8).slice(1);
        },
        decimal: function(obj) {
          return obj.toString(10);
        },
        hexadecimal: function(obj) {
          return obj >= 0 ? "0x" + obj.toString(16).toUpperCase() : "-0x" + obj.toString(16).toUpperCase().slice(1);
        }
      },
      defaultStyle: "decimal",
      styleAliases: {
        binary: [2, "bin"],
        octal: [8, "oct"],
        decimal: [10, "dec"],
        hexadecimal: [16, "hex"]
      }
    });
    return int;
  }
  var float;
  var hasRequiredFloat;
  function requireFloat() {
    if (hasRequiredFloat) return float;
    hasRequiredFloat = 1;
    const common2 = requireCommon();
    const Type2 = requireType();
    const YAML_FLOAT_PATTERN = new RegExp(
      // 2.5e4, 2.5 and integers
      "^(?:[-+]?(?:[0-9]+)(?:\\.[0-9]*)?(?:[eE][-+]?[0-9]+)?|\\.[0-9]+(?:[eE][-+]?[0-9]+)?|[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"
    );
    const YAML_FLOAT_SPECIAL_PATTERN = new RegExp(
      "^(?:[-+]?\\.(?:inf|Inf|INF)|\\.(?:nan|NaN|NAN))$"
    );
    function resolveYamlFloat(data) {
      if (data === null) return false;
      if (!YAML_FLOAT_PATTERN.test(data)) {
        return false;
      }
      if (isFinite(parseFloat(data, 10))) {
        return true;
      }
      return YAML_FLOAT_SPECIAL_PATTERN.test(data);
    }
    function constructYamlFloat(data) {
      let value = data.toLowerCase();
      const sign = value[0] === "-" ? -1 : 1;
      if ("+-".indexOf(value[0]) >= 0) {
        value = value.slice(1);
      }
      if (value === ".inf") {
        return sign === 1 ? Number.POSITIVE_INFINITY : Number.NEGATIVE_INFINITY;
      } else if (value === ".nan") {
        return NaN;
      }
      return sign * parseFloat(value, 10);
    }
    const SCIENTIFIC_WITHOUT_DOT = /^[-+]?[0-9]+e/;
    function representYamlFloat(object, style) {
      if (isNaN(object)) {
        switch (style) {
          case "lowercase":
            return ".nan";
          case "uppercase":
            return ".NAN";
          case "camelcase":
            return ".NaN";
        }
      } else if (Number.POSITIVE_INFINITY === object) {
        switch (style) {
          case "lowercase":
            return ".inf";
          case "uppercase":
            return ".INF";
          case "camelcase":
            return ".Inf";
        }
      } else if (Number.NEGATIVE_INFINITY === object) {
        switch (style) {
          case "lowercase":
            return "-.inf";
          case "uppercase":
            return "-.INF";
          case "camelcase":
            return "-.Inf";
        }
      } else if (common2.isNegativeZero(object)) {
        return "-0.0";
      }
      const res = object.toString(10);
      return SCIENTIFIC_WITHOUT_DOT.test(res) ? res.replace("e", ".e") : res;
    }
    function isFloat(object) {
      return Object.prototype.toString.call(object) === "[object Number]" && (object % 1 !== 0 || common2.isNegativeZero(object));
    }
    float = new Type2("tag:yaml.org,2002:float", {
      kind: "scalar",
      resolve: resolveYamlFloat,
      construct: constructYamlFloat,
      predicate: isFloat,
      represent: representYamlFloat,
      defaultStyle: "lowercase"
    });
    return float;
  }
  var json;
  var hasRequiredJson;
  function requireJson() {
    if (hasRequiredJson) return json;
    hasRequiredJson = 1;
    json = requireFailsafe().extend({
      implicit: [
        require_null(),
        requireBool(),
        requireInt(),
        requireFloat()
      ]
    });
    return json;
  }
  var core;
  var hasRequiredCore;
  function requireCore() {
    if (hasRequiredCore) return core;
    hasRequiredCore = 1;
    core = requireJson();
    return core;
  }
  var timestamp;
  var hasRequiredTimestamp;
  function requireTimestamp() {
    if (hasRequiredTimestamp) return timestamp;
    hasRequiredTimestamp = 1;
    const Type2 = requireType();
    const YAML_DATE_REGEXP = new RegExp(
      "^([0-9][0-9][0-9][0-9])-([0-9][0-9])-([0-9][0-9])$"
    );
    const YAML_TIMESTAMP_REGEXP = new RegExp(
      "^([0-9][0-9][0-9][0-9])-([0-9][0-9]?)-([0-9][0-9]?)(?:[Tt]|[ \\t]+)([0-9][0-9]?):([0-9][0-9]):([0-9][0-9])(?:\\.([0-9]*))?(?:[ \\t]*(Z|([-+])([0-9][0-9]?)(?::([0-9][0-9]))?))?$"
    );
    function resolveYamlTimestamp(data) {
      if (data === null) return false;
      if (YAML_DATE_REGEXP.exec(data) !== null) return true;
      if (YAML_TIMESTAMP_REGEXP.exec(data) !== null) return true;
      return false;
    }
    function constructYamlTimestamp(data) {
      let fraction = 0;
      let delta = null;
      let match = YAML_DATE_REGEXP.exec(data);
      if (match === null) match = YAML_TIMESTAMP_REGEXP.exec(data);
      if (match === null) throw new Error("Date resolve error");
      const year = +match[1];
      const month = +match[2] - 1;
      const day = +match[3];
      if (!match[4]) {
        return new Date(Date.UTC(year, month, day));
      }
      const hour = +match[4];
      const minute = +match[5];
      const second = +match[6];
      if (match[7]) {
        fraction = match[7].slice(0, 3);
        while (fraction.length < 3) {
          fraction += "0";
        }
        fraction = +fraction;
      }
      if (match[9]) {
        const tzHour = +match[10];
        const tzMinute = +(match[11] || 0);
        delta = (tzHour * 60 + tzMinute) * 6e4;
        if (match[9] === "-") delta = -delta;
      }
      const date = new Date(Date.UTC(year, month, day, hour, minute, second, fraction));
      if (delta) date.setTime(date.getTime() - delta);
      return date;
    }
    function representYamlTimestamp(object) {
      return object.toISOString();
    }
    timestamp = new Type2("tag:yaml.org,2002:timestamp", {
      kind: "scalar",
      resolve: resolveYamlTimestamp,
      construct: constructYamlTimestamp,
      instanceOf: Date,
      represent: representYamlTimestamp
    });
    return timestamp;
  }
  var merge;
  var hasRequiredMerge;
  function requireMerge() {
    if (hasRequiredMerge) return merge;
    hasRequiredMerge = 1;
    const Type2 = requireType();
    function resolveYamlMerge(data) {
      return data === "<<" || data === null;
    }
    merge = new Type2("tag:yaml.org,2002:merge", {
      kind: "scalar",
      resolve: resolveYamlMerge
    });
    return merge;
  }
  var binary;
  var hasRequiredBinary;
  function requireBinary() {
    if (hasRequiredBinary) return binary;
    hasRequiredBinary = 1;
    const Type2 = requireType();
    const BASE64_MAP = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\n\r";
    function resolveYamlBinary(data) {
      if (data === null) return false;
      let bitlen = 0;
      const max = data.length;
      const map2 = BASE64_MAP;
      for (let idx = 0; idx < max; idx++) {
        const code = map2.indexOf(data.charAt(idx));
        if (code > 64) continue;
        if (code < 0) return false;
        bitlen += 6;
      }
      return bitlen % 8 === 0;
    }
    function constructYamlBinary(data) {
      const input = data.replace(/[\r\n=]/g, "");
      const max = input.length;
      const map2 = BASE64_MAP;
      let bits = 0;
      const result = [];
      for (let idx = 0; idx < max; idx++) {
        if (idx % 4 === 0 && idx) {
          result.push(bits >> 16 & 255);
          result.push(bits >> 8 & 255);
          result.push(bits & 255);
        }
        bits = bits << 6 | map2.indexOf(input.charAt(idx));
      }
      const tailbits = max % 4 * 6;
      if (tailbits === 0) {
        result.push(bits >> 16 & 255);
        result.push(bits >> 8 & 255);
        result.push(bits & 255);
      } else if (tailbits === 18) {
        result.push(bits >> 10 & 255);
        result.push(bits >> 2 & 255);
      } else if (tailbits === 12) {
        result.push(bits >> 4 & 255);
      }
      return new Uint8Array(result);
    }
    function representYamlBinary(object) {
      let result = "";
      let bits = 0;
      const max = object.length;
      const map2 = BASE64_MAP;
      for (let idx = 0; idx < max; idx++) {
        if (idx % 3 === 0 && idx) {
          result += map2[bits >> 18 & 63];
          result += map2[bits >> 12 & 63];
          result += map2[bits >> 6 & 63];
          result += map2[bits & 63];
        }
        bits = (bits << 8) + object[idx];
      }
      const tail = max % 3;
      if (tail === 0) {
        result += map2[bits >> 18 & 63];
        result += map2[bits >> 12 & 63];
        result += map2[bits >> 6 & 63];
        result += map2[bits & 63];
      } else if (tail === 2) {
        result += map2[bits >> 10 & 63];
        result += map2[bits >> 4 & 63];
        result += map2[bits << 2 & 63];
        result += map2[64];
      } else if (tail === 1) {
        result += map2[bits >> 2 & 63];
        result += map2[bits << 4 & 63];
        result += map2[64];
        result += map2[64];
      }
      return result;
    }
    function isBinary(obj) {
      return Object.prototype.toString.call(obj) === "[object Uint8Array]";
    }
    binary = new Type2("tag:yaml.org,2002:binary", {
      kind: "scalar",
      resolve: resolveYamlBinary,
      construct: constructYamlBinary,
      predicate: isBinary,
      represent: representYamlBinary
    });
    return binary;
  }
  var omap;
  var hasRequiredOmap;
  function requireOmap() {
    if (hasRequiredOmap) return omap;
    hasRequiredOmap = 1;
    const Type2 = requireType();
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    const _toString = Object.prototype.toString;
    function resolveYamlOmap(data) {
      if (data === null) return true;
      const objectKeys2 = {};
      const object = data;
      for (let index = 0, length = object.length; index < length; index += 1) {
        const pair = object[index];
        let pairHasKey = false;
        if (_toString.call(pair) !== "[object Object]") return false;
        let pairKey;
        for (pairKey in pair) {
          if (_hasOwnProperty2.call(pair, pairKey)) {
            if (!pairHasKey) pairHasKey = true;
            else return false;
          }
        }
        if (!pairHasKey) return false;
        if (_hasOwnProperty2.call(objectKeys2, pairKey)) return false;
        Object.defineProperty(objectKeys2, pairKey, { value: true });
      }
      return true;
    }
    function constructYamlOmap(data) {
      return data !== null ? data : [];
    }
    omap = new Type2("tag:yaml.org,2002:omap", {
      kind: "sequence",
      resolve: resolveYamlOmap,
      construct: constructYamlOmap
    });
    return omap;
  }
  var pairs;
  var hasRequiredPairs;
  function requirePairs() {
    if (hasRequiredPairs) return pairs;
    hasRequiredPairs = 1;
    const Type2 = requireType();
    const _toString = Object.prototype.toString;
    function resolveYamlPairs(data) {
      if (data === null) return true;
      const object = data;
      const result = new Array(object.length);
      for (let index = 0, length = object.length; index < length; index += 1) {
        const pair = object[index];
        if (_toString.call(pair) !== "[object Object]") return false;
        const keys = Object.keys(pair);
        if (keys.length !== 1) return false;
        result[index] = [keys[0], pair[keys[0]]];
      }
      return true;
    }
    function constructYamlPairs(data) {
      if (data === null) return [];
      const object = data;
      const result = new Array(object.length);
      for (let index = 0, length = object.length; index < length; index += 1) {
        const pair = object[index];
        const keys = Object.keys(pair);
        result[index] = [keys[0], pair[keys[0]]];
      }
      return result;
    }
    pairs = new Type2("tag:yaml.org,2002:pairs", {
      kind: "sequence",
      resolve: resolveYamlPairs,
      construct: constructYamlPairs
    });
    return pairs;
  }
  var set;
  var hasRequiredSet;
  function requireSet() {
    if (hasRequiredSet) return set;
    hasRequiredSet = 1;
    const Type2 = requireType();
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    function resolveYamlSet(data) {
      if (data === null) return true;
      const object = data;
      for (const key in object) {
        if (_hasOwnProperty2.call(object, key)) {
          if (object[key] !== null) return false;
        }
      }
      return true;
    }
    function constructYamlSet(data) {
      return data !== null ? data : {};
    }
    set = new Type2("tag:yaml.org,2002:set", {
      kind: "mapping",
      resolve: resolveYamlSet,
      construct: constructYamlSet
    });
    return set;
  }
  var _default;
  var hasRequired_default;
  function require_default() {
    if (hasRequired_default) return _default;
    hasRequired_default = 1;
    _default = requireCore().extend({
      implicit: [
        requireTimestamp(),
        requireMerge()
      ],
      explicit: [
        requireBinary(),
        requireOmap(),
        requirePairs(),
        requireSet()
      ]
    });
    return _default;
  }
  var hasRequiredLoader;
  function requireLoader() {
    if (hasRequiredLoader) return loader;
    hasRequiredLoader = 1;
    const common2 = requireCommon();
    const YAMLException2 = requireException();
    const makeSnippet = requireSnippet();
    const DEFAULT_SCHEMA2 = require_default();
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    const CONTEXT_FLOW_IN = 1;
    const CONTEXT_FLOW_OUT = 2;
    const CONTEXT_BLOCK_IN = 3;
    const CONTEXT_BLOCK_OUT = 4;
    const CHOMPING_CLIP = 1;
    const CHOMPING_STRIP = 2;
    const CHOMPING_KEEP = 3;
    const PATTERN_NON_PRINTABLE = /[\x00-\x08\x0B\x0C\x0E-\x1F\x7F-\x84\x86-\x9F\uFFFE\uFFFF]|[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?:[^\uD800-\uDBFF]|^)[\uDC00-\uDFFF]/;
    const PATTERN_NON_ASCII_LINE_BREAKS = /[\x85\u2028\u2029]/;
    const PATTERN_FLOW_INDICATORS = /[,\[\]{}]/;
    const PATTERN_TAG_HANDLE = /^(?:!|!!|![0-9A-Za-z-]+!)$/;
    const PATTERN_TAG_URI = /^(?:!|[^,\[\]{}])(?:%[0-9a-f]{2}|[0-9a-z\-#;/?:@&=+$,_.!~*'()\[\]])*$/i;
    function _class(obj) {
      return Object.prototype.toString.call(obj);
    }
    function isEol(c) {
      return c === 10 || c === 13;
    }
    function isWhiteSpace(c) {
      return c === 9 || c === 32;
    }
    function isWsOrEol(c) {
      return c === 9 || c === 32 || c === 10 || c === 13;
    }
    function isFlowIndicator(c) {
      return c === 44 || c === 91 || c === 93 || c === 123 || c === 125;
    }
    function fromHexCode(c) {
      if (c >= 48 && c <= 57) {
        return c - 48;
      }
      const lc = c | 32;
      if (lc >= 97 && lc <= 102) {
        return lc - 97 + 10;
      }
      return -1;
    }
    function escapedHexLen(c) {
      if (c === 120) {
        return 2;
      }
      if (c === 117) {
        return 4;
      }
      if (c === 85) {
        return 8;
      }
      return 0;
    }
    function fromDecimalCode(c) {
      if (c >= 48 && c <= 57) {
        return c - 48;
      }
      return -1;
    }
    function simpleEscapeSequence(c) {
      switch (c) {
        case 48:
          return "\0";
        case 97:
          return "\x07";
        case 98:
          return "\b";
        case 116:
          return "	";
        case 9:
          return "	";
        case 110:
          return "\n";
        case 118:
          return "\v";
        case 102:
          return "\f";
        case 114:
          return "\r";
        case 101:
          return "\x1B";
        case 32:
          return " ";
        case 34:
          return '"';
        case 47:
          return "/";
        case 92:
          return "\\";
        case 78:
          return "\x85";
        case 95:
          return "\xA0";
        case 76:
          return "\u2028";
        case 80:
          return "\u2029";
        default:
          return "";
      }
    }
    function charFromCodepoint(c) {
      if (c <= 65535) {
        return String.fromCharCode(c);
      }
      return String.fromCharCode(
        (c - 65536 >> 10) + 55296,
        (c - 65536 & 1023) + 56320
      );
    }
    function setProperty(object, key, value) {
      if (key === "__proto__") {
        Object.defineProperty(object, key, {
          configurable: true,
          enumerable: true,
          writable: true,
          value
        });
      } else {
        object[key] = value;
      }
    }
    const simpleEscapeCheck = new Array(256);
    const simpleEscapeMap = new Array(256);
    for (let i = 0; i < 256; i++) {
      simpleEscapeCheck[i] = simpleEscapeSequence(i) ? 1 : 0;
      simpleEscapeMap[i] = simpleEscapeSequence(i);
    }
    function State(input, options2) {
      this.input = input;
      this.filename = options2["filename"] || null;
      this.schema = options2["schema"] || DEFAULT_SCHEMA2;
      this.onWarning = options2["onWarning"] || null;
      this.legacy = options2["legacy"] || false;
      this.json = options2["json"] || false;
      this.listener = options2["listener"] || null;
      this.maxDepth = typeof options2["maxDepth"] === "number" ? options2["maxDepth"] : 100;
      this.maxTotalMergeKeys = typeof options2["maxTotalMergeKeys"] === "number" ? options2["maxTotalMergeKeys"] : 1e4;
      this.implicitTypes = this.schema.compiledImplicit;
      this.typeMap = this.schema.compiledTypeMap;
      this.length = input.length;
      this.position = 0;
      this.line = 0;
      this.lineStart = 0;
      this.lineIndent = 0;
      this.depth = 0;
      this.totalMergeKeys = 0;
      this.firstTabInLine = -1;
      this.documents = [];
      this.anchorMapTransactions = [];
    }
    function generateError(state, message) {
      const mark = {
        name: state.filename,
        buffer: state.input.slice(0, -1),
        // omit trailing \0
        position: state.position,
        line: state.line,
        column: state.position - state.lineStart
      };
      mark.snippet = makeSnippet(mark);
      return new YAMLException2(message, mark);
    }
    function throwError(state, message) {
      throw generateError(state, message);
    }
    function throwWarning(state, message) {
      if (state.onWarning) {
        state.onWarning.call(null, generateError(state, message));
      }
    }
    function storeAnchor(state, name, value) {
      const transactions = state.anchorMapTransactions;
      if (transactions.length !== 0) {
        const transaction = transactions[transactions.length - 1];
        if (!_hasOwnProperty2.call(transaction, name)) {
          transaction[name] = {
            existed: _hasOwnProperty2.call(state.anchorMap, name),
            value: state.anchorMap[name]
          };
        }
      }
      state.anchorMap[name] = value;
    }
    function beginAnchorTransaction(state) {
      state.anchorMapTransactions.push(/* @__PURE__ */ Object.create(null));
    }
    function commitAnchorTransaction(state) {
      const transaction = state.anchorMapTransactions.pop();
      const transactions = state.anchorMapTransactions;
      if (transactions.length === 0) return;
      const parent = transactions[transactions.length - 1];
      const names = Object.keys(transaction);
      for (let index = 0, length = names.length; index < length; index += 1) {
        const name = names[index];
        if (!_hasOwnProperty2.call(parent, name)) {
          parent[name] = transaction[name];
        }
      }
    }
    function rollbackAnchorTransaction(state) {
      const transaction = state.anchorMapTransactions.pop();
      const names = Object.keys(transaction);
      for (let index = names.length - 1; index >= 0; index -= 1) {
        const entry = transaction[names[index]];
        if (entry.existed) {
          state.anchorMap[names[index]] = entry.value;
        } else {
          delete state.anchorMap[names[index]];
        }
      }
    }
    function snapshotState(state) {
      return {
        position: state.position,
        line: state.line,
        lineStart: state.lineStart,
        lineIndent: state.lineIndent,
        firstTabInLine: state.firstTabInLine,
        tag: state.tag,
        anchor: state.anchor,
        kind: state.kind,
        result: state.result
      };
    }
    function restoreState(state, snapshot) {
      state.position = snapshot.position;
      state.line = snapshot.line;
      state.lineStart = snapshot.lineStart;
      state.lineIndent = snapshot.lineIndent;
      state.firstTabInLine = snapshot.firstTabInLine;
      state.tag = snapshot.tag;
      state.anchor = snapshot.anchor;
      state.kind = snapshot.kind;
      state.result = snapshot.result;
    }
    const directiveHandlers = {
      YAML: function handleYamlDirective(state, name, args) {
        if (state.version !== null) {
          throwError(state, "duplication of %YAML directive");
        }
        if (args.length !== 1) {
          throwError(state, "YAML directive accepts exactly one argument");
        }
        const match = /^([0-9]+)\.([0-9]+)$/.exec(args[0]);
        if (match === null) {
          throwError(state, "ill-formed argument of the YAML directive");
        }
        const major = parseInt(match[1], 10);
        const minor = parseInt(match[2], 10);
        if (major !== 1) {
          throwError(state, "unacceptable YAML version of the document");
        }
        state.version = args[0];
        state.checkLineBreaks = minor < 2;
        if (minor !== 1 && minor !== 2) {
          throwWarning(state, "unsupported YAML version of the document");
        }
      },
      TAG: function handleTagDirective(state, name, args) {
        let prefix;
        if (args.length !== 2) {
          throwError(state, "TAG directive accepts exactly two arguments");
        }
        const handle = args[0];
        prefix = args[1];
        if (!PATTERN_TAG_HANDLE.test(handle)) {
          throwError(state, "ill-formed tag handle (first argument) of the TAG directive");
        }
        if (_hasOwnProperty2.call(state.tagMap, handle)) {
          throwError(state, 'there is a previously declared suffix for "' + handle + '" tag handle');
        }
        if (!PATTERN_TAG_URI.test(prefix)) {
          throwError(state, "ill-formed tag prefix (second argument) of the TAG directive");
        }
        try {
          prefix = decodeURIComponent(prefix);
        } catch (err) {
          throwError(state, "tag prefix is malformed: " + prefix);
        }
        state.tagMap[handle] = prefix;
      }
    };
    function captureSegment(state, start, end, checkJson) {
      if (start < end) {
        const _result = state.input.slice(start, end);
        if (checkJson) {
          for (let _position = 0, _length = _result.length; _position < _length; _position += 1) {
            const _character = _result.charCodeAt(_position);
            if (!(_character === 9 || _character >= 32 && _character <= 1114111)) {
              throwError(state, "expected valid JSON character");
            }
          }
        } else if (PATTERN_NON_PRINTABLE.test(_result)) {
          throwError(state, "the stream contains non-printable characters");
        }
        state.result += _result;
      }
    }
    function chargeMergeWork(state) {
      state.totalMergeKeys++;
      if (state.maxTotalMergeKeys !== -1 && state.totalMergeKeys > state.maxTotalMergeKeys) {
        throwError(state, "merge keys exceeded maxTotalMergeKeys (" + state.maxTotalMergeKeys + ")");
      }
    }
    function mergeMappings(state, destination, source, overridableKeys) {
      if (!common2.isObject(source)) {
        throwError(state, "cannot merge mappings; the provided source object is unacceptable");
      }
      chargeMergeWork(state);
      const sourceKeys = Object.keys(source);
      for (let index = 0, quantity = sourceKeys.length; index < quantity; index += 1) {
        const key = sourceKeys[index];
        chargeMergeWork(state);
        if (!_hasOwnProperty2.call(destination, key)) {
          setProperty(destination, key, source[key]);
          overridableKeys[key] = true;
        }
      }
    }
    function storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, startLine, startLineStart, startPos) {
      if (Array.isArray(keyNode)) {
        keyNode = Array.prototype.slice.call(keyNode);
        for (let index = 0, quantity = keyNode.length; index < quantity; index += 1) {
          if (Array.isArray(keyNode[index])) {
            throwError(state, "nested arrays are not supported inside keys");
          }
          if (typeof keyNode === "object" && _class(keyNode[index]) === "[object Object]") {
            keyNode[index] = "[object Object]";
          }
        }
      }
      if (typeof keyNode === "object" && _class(keyNode) === "[object Object]") {
        keyNode = "[object Object]";
      }
      keyNode = String(keyNode);
      if (_result === null) {
        _result = {};
      }
      if (keyTag === "tag:yaml.org,2002:merge") {
        if (Array.isArray(valueNode)) {
          if (valueNode.length > 100) {
            throwError(state, "abnormal merge sequence size");
          }
          for (let index = 0, quantity = valueNode.length; index < quantity; index += 1) {
            mergeMappings(state, _result, valueNode[index], overridableKeys);
          }
        } else {
          mergeMappings(state, _result, valueNode, overridableKeys);
        }
      } else {
        if (!state.json && !_hasOwnProperty2.call(overridableKeys, keyNode) && _hasOwnProperty2.call(_result, keyNode)) {
          state.line = startLine || state.line;
          state.lineStart = startLineStart || state.lineStart;
          state.position = startPos || state.position;
          throwError(state, "duplicated mapping key");
        }
        setProperty(_result, keyNode, valueNode);
        delete overridableKeys[keyNode];
      }
      return _result;
    }
    function readLineBreak(state) {
      const ch = state.input.charCodeAt(state.position);
      if (ch === 10) {
        state.position++;
      } else if (ch === 13) {
        state.position++;
        if (state.input.charCodeAt(state.position) === 10) {
          state.position++;
        }
      } else {
        throwError(state, "a line break is expected");
      }
      state.line += 1;
      state.lineStart = state.position;
      state.firstTabInLine = -1;
    }
    function skipSeparationSpace(state, allowComments, checkIndent) {
      let lineBreaks = 0;
      let ch = state.input.charCodeAt(state.position);
      while (ch !== 0) {
        while (isWhiteSpace(ch)) {
          if (ch === 9 && state.firstTabInLine === -1) {
            state.firstTabInLine = state.position;
          }
          ch = state.input.charCodeAt(++state.position);
        }
        if (allowComments && ch === 35) {
          do {
            ch = state.input.charCodeAt(++state.position);
          } while (ch !== 10 && ch !== 13 && ch !== 0);
        }
        if (isEol(ch)) {
          readLineBreak(state);
          ch = state.input.charCodeAt(state.position);
          lineBreaks++;
          state.lineIndent = 0;
          while (ch === 32) {
            state.lineIndent++;
            ch = state.input.charCodeAt(++state.position);
          }
        } else {
          break;
        }
      }
      if (checkIndent !== -1 && lineBreaks !== 0 && state.lineIndent < checkIndent) {
        throwWarning(state, "deficient indentation");
      }
      return lineBreaks;
    }
    function testDocumentSeparator(state) {
      let _position = state.position;
      let ch = state.input.charCodeAt(_position);
      if ((ch === 45 || ch === 46) && ch === state.input.charCodeAt(_position + 1) && ch === state.input.charCodeAt(_position + 2)) {
        _position += 3;
        ch = state.input.charCodeAt(_position);
        if (ch === 0 || isWsOrEol(ch)) {
          return true;
        }
      }
      return false;
    }
    function writeFoldedLines(state, count) {
      if (count === 1) {
        state.result += " ";
      } else if (count > 1) {
        state.result += common2.repeat("\n", count - 1);
      }
    }
    function readPlainScalar(state, nodeIndent, withinFlowCollection) {
      let captureStart;
      let captureEnd;
      let hasPendingContent;
      let _line;
      let _lineStart;
      let _lineIndent;
      const _kind = state.kind;
      const _result = state.result;
      let ch = state.input.charCodeAt(state.position);
      if (isWsOrEol(ch) || isFlowIndicator(ch) || ch === 35 || ch === 38 || ch === 42 || ch === 33 || ch === 124 || ch === 62 || ch === 39 || ch === 34 || ch === 37 || ch === 64 || ch === 96) {
        return false;
      }
      if (ch === 63 || ch === 45) {
        const following = state.input.charCodeAt(state.position + 1);
        if (isWsOrEol(following) || withinFlowCollection && isFlowIndicator(following)) {
          return false;
        }
      }
      state.kind = "scalar";
      state.result = "";
      captureStart = captureEnd = state.position;
      hasPendingContent = false;
      while (ch !== 0) {
        if (ch === 58) {
          const following = state.input.charCodeAt(state.position + 1);
          if (isWsOrEol(following) || withinFlowCollection && isFlowIndicator(following)) {
            break;
          }
        } else if (ch === 35) {
          const preceding = state.input.charCodeAt(state.position - 1);
          if (isWsOrEol(preceding)) {
            break;
          }
        } else if (state.position === state.lineStart && testDocumentSeparator(state) || withinFlowCollection && isFlowIndicator(ch)) {
          break;
        } else if (isEol(ch)) {
          _line = state.line;
          _lineStart = state.lineStart;
          _lineIndent = state.lineIndent;
          skipSeparationSpace(state, false, -1);
          if (state.lineIndent >= nodeIndent) {
            hasPendingContent = true;
            ch = state.input.charCodeAt(state.position);
            continue;
          } else {
            state.position = captureEnd;
            state.line = _line;
            state.lineStart = _lineStart;
            state.lineIndent = _lineIndent;
            break;
          }
        }
        if (hasPendingContent) {
          captureSegment(state, captureStart, captureEnd, false);
          writeFoldedLines(state, state.line - _line);
          captureStart = captureEnd = state.position;
          hasPendingContent = false;
        }
        if (!isWhiteSpace(ch)) {
          captureEnd = state.position + 1;
        }
        ch = state.input.charCodeAt(++state.position);
      }
      captureSegment(state, captureStart, captureEnd, false);
      if (state.result) {
        return true;
      }
      state.kind = _kind;
      state.result = _result;
      return false;
    }
    function readSingleQuotedScalar(state, nodeIndent) {
      let captureStart;
      let captureEnd;
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 39) {
        return false;
      }
      state.kind = "scalar";
      state.result = "";
      state.position++;
      captureStart = captureEnd = state.position;
      while ((ch = state.input.charCodeAt(state.position)) !== 0) {
        if (ch === 39) {
          captureSegment(state, captureStart, state.position, true);
          ch = state.input.charCodeAt(++state.position);
          if (ch === 39) {
            captureStart = state.position;
            state.position++;
            captureEnd = state.position;
          } else {
            return true;
          }
        } else if (isEol(ch)) {
          captureSegment(state, captureStart, captureEnd, true);
          writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
          captureStart = captureEnd = state.position;
        } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
          throwError(state, "unexpected end of the document within a single quoted scalar");
        } else {
          state.position++;
          if (!isWhiteSpace(ch)) {
            captureEnd = state.position;
          }
        }
      }
      throwError(state, "unexpected end of the stream within a single quoted scalar");
    }
    function readDoubleQuotedScalar(state, nodeIndent) {
      let captureStart;
      let captureEnd;
      let tmp;
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 34) {
        return false;
      }
      state.kind = "scalar";
      state.result = "";
      state.position++;
      captureStart = captureEnd = state.position;
      while ((ch = state.input.charCodeAt(state.position)) !== 0) {
        if (ch === 34) {
          captureSegment(state, captureStart, state.position, true);
          state.position++;
          return true;
        } else if (ch === 92) {
          captureSegment(state, captureStart, state.position, true);
          ch = state.input.charCodeAt(++state.position);
          if (isEol(ch)) {
            skipSeparationSpace(state, false, nodeIndent);
          } else if (ch < 256 && simpleEscapeCheck[ch]) {
            state.result += simpleEscapeMap[ch];
            state.position++;
          } else if ((tmp = escapedHexLen(ch)) > 0) {
            let hexLength = tmp;
            let hexResult = 0;
            for (; hexLength > 0; hexLength--) {
              ch = state.input.charCodeAt(++state.position);
              if ((tmp = fromHexCode(ch)) >= 0) {
                hexResult = (hexResult << 4) + tmp;
              } else {
                throwError(state, "expected hexadecimal character");
              }
            }
            state.result += charFromCodepoint(hexResult);
            state.position++;
          } else {
            throwError(state, "unknown escape sequence");
          }
          captureStart = captureEnd = state.position;
        } else if (isEol(ch)) {
          captureSegment(state, captureStart, captureEnd, true);
          writeFoldedLines(state, skipSeparationSpace(state, false, nodeIndent));
          captureStart = captureEnd = state.position;
        } else if (state.position === state.lineStart && testDocumentSeparator(state)) {
          throwError(state, "unexpected end of the document within a double quoted scalar");
        } else {
          state.position++;
          if (!isWhiteSpace(ch)) {
            captureEnd = state.position;
          }
        }
      }
      throwError(state, "unexpected end of the stream within a double quoted scalar");
    }
    function readFlowCollection(state, nodeIndent) {
      let readNext = true;
      let _line;
      let _lineStart;
      let _pos;
      const _tag2 = state.tag;
      let _result;
      const _anchor = state.anchor;
      let terminator;
      let isPair;
      let isExplicitPair;
      let isMapping;
      const overridableKeys = /* @__PURE__ */ Object.create(null);
      let keyNode;
      let keyTag;
      let valueNode;
      let ch = state.input.charCodeAt(state.position);
      if (ch === 91) {
        terminator = 93;
        isMapping = false;
        _result = [];
      } else if (ch === 123) {
        terminator = 125;
        isMapping = true;
        _result = {};
      } else {
        return false;
      }
      if (state.anchor !== null) {
        storeAnchor(state, state.anchor, _result);
      }
      ch = state.input.charCodeAt(++state.position);
      while (ch !== 0) {
        skipSeparationSpace(state, true, nodeIndent);
        ch = state.input.charCodeAt(state.position);
        if (ch === terminator) {
          state.position++;
          state.tag = _tag2;
          state.anchor = _anchor;
          state.kind = isMapping ? "mapping" : "sequence";
          state.result = _result;
          return true;
        } else if (!readNext) {
          throwError(state, "missed comma between flow collection entries");
        } else if (ch === 44) {
          throwError(state, "expected the node content, but found ','");
        }
        keyTag = keyNode = valueNode = null;
        isPair = isExplicitPair = false;
        if (ch === 63) {
          const following = state.input.charCodeAt(state.position + 1);
          if (isWsOrEol(following)) {
            isPair = isExplicitPair = true;
            state.position++;
            skipSeparationSpace(state, true, nodeIndent);
          }
        }
        _line = state.line;
        _lineStart = state.lineStart;
        _pos = state.position;
        composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
        keyTag = state.tag;
        keyNode = state.result;
        skipSeparationSpace(state, true, nodeIndent);
        ch = state.input.charCodeAt(state.position);
        if ((isExplicitPair || state.line === _line) && ch === 58) {
          isPair = true;
          ch = state.input.charCodeAt(++state.position);
          skipSeparationSpace(state, true, nodeIndent);
          composeNode(state, nodeIndent, CONTEXT_FLOW_IN, false, true);
          valueNode = state.result;
        }
        if (isMapping) {
          storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos);
        } else if (isPair) {
          _result.push(storeMappingPair(state, null, overridableKeys, keyTag, keyNode, valueNode, _line, _lineStart, _pos));
        } else {
          _result.push(keyNode);
        }
        skipSeparationSpace(state, true, nodeIndent);
        ch = state.input.charCodeAt(state.position);
        if (ch === 44) {
          readNext = true;
          ch = state.input.charCodeAt(++state.position);
        } else {
          readNext = false;
        }
      }
      throwError(state, "unexpected end of the stream within a flow collection");
    }
    function readBlockScalar(state, nodeIndent) {
      let folding;
      let chomping = CHOMPING_CLIP;
      let didReadContent = false;
      let detectedIndent = false;
      let textIndent = nodeIndent;
      let emptyLines = 0;
      let atMoreIndented = false;
      let tmp;
      let ch = state.input.charCodeAt(state.position);
      if (ch === 124) {
        folding = false;
      } else if (ch === 62) {
        folding = true;
      } else {
        return false;
      }
      state.kind = "scalar";
      state.result = "";
      while (ch !== 0) {
        ch = state.input.charCodeAt(++state.position);
        if (ch === 43 || ch === 45) {
          if (CHOMPING_CLIP === chomping) {
            chomping = ch === 43 ? CHOMPING_KEEP : CHOMPING_STRIP;
          } else {
            throwError(state, "repeat of a chomping mode identifier");
          }
        } else if ((tmp = fromDecimalCode(ch)) >= 0) {
          if (tmp === 0) {
            throwError(state, "bad explicit indentation width of a block scalar; it cannot be less than one");
          } else if (!detectedIndent) {
            textIndent = nodeIndent + tmp - 1;
            detectedIndent = true;
          } else {
            throwError(state, "repeat of an indentation width identifier");
          }
        } else {
          break;
        }
      }
      if (isWhiteSpace(ch)) {
        do {
          ch = state.input.charCodeAt(++state.position);
        } while (isWhiteSpace(ch));
        if (ch === 35) {
          do {
            ch = state.input.charCodeAt(++state.position);
          } while (!isEol(ch) && ch !== 0);
        }
      }
      while (ch !== 0) {
        readLineBreak(state);
        state.lineIndent = 0;
        ch = state.input.charCodeAt(state.position);
        while ((!detectedIndent || state.lineIndent < textIndent) && ch === 32) {
          state.lineIndent++;
          ch = state.input.charCodeAt(++state.position);
        }
        if (!detectedIndent && state.lineIndent > textIndent) {
          textIndent = state.lineIndent;
        }
        if (isEol(ch)) {
          emptyLines++;
          continue;
        }
        if (!detectedIndent && textIndent === 0) {
          throwError(state, "missing indentation for block scalar");
        }
        if (state.lineIndent < textIndent) {
          if (chomping === CHOMPING_KEEP) {
            state.result += common2.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
          } else if (chomping === CHOMPING_CLIP) {
            if (didReadContent) {
              state.result += "\n";
            }
          }
          break;
        }
        if (folding) {
          if (isWhiteSpace(ch)) {
            atMoreIndented = true;
            state.result += common2.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
          } else if (atMoreIndented) {
            atMoreIndented = false;
            state.result += common2.repeat("\n", emptyLines + 1);
          } else if (emptyLines === 0) {
            if (didReadContent) {
              state.result += " ";
            }
          } else {
            state.result += common2.repeat("\n", emptyLines);
          }
        } else {
          state.result += common2.repeat("\n", didReadContent ? 1 + emptyLines : emptyLines);
        }
        didReadContent = true;
        detectedIndent = true;
        emptyLines = 0;
        const captureStart = state.position;
        while (!isEol(ch) && ch !== 0) {
          ch = state.input.charCodeAt(++state.position);
        }
        captureSegment(state, captureStart, state.position, false);
      }
      return true;
    }
    function readBlockSequence(state, nodeIndent) {
      const _tag2 = state.tag;
      const _anchor = state.anchor;
      const _result = [];
      let detected = false;
      if (state.firstTabInLine !== -1) return false;
      if (state.anchor !== null) {
        storeAnchor(state, state.anchor, _result);
      }
      let ch = state.input.charCodeAt(state.position);
      while (ch !== 0) {
        if (state.firstTabInLine !== -1) {
          state.position = state.firstTabInLine;
          throwError(state, "tab characters must not be used in indentation");
        }
        if (ch !== 45) {
          break;
        }
        const following = state.input.charCodeAt(state.position + 1);
        if (!isWsOrEol(following)) {
          break;
        }
        detected = true;
        state.position++;
        if (skipSeparationSpace(state, true, -1)) {
          if (state.lineIndent <= nodeIndent) {
            _result.push(null);
            ch = state.input.charCodeAt(state.position);
            continue;
          }
        }
        const _line = state.line;
        composeNode(state, nodeIndent, CONTEXT_BLOCK_IN, false, true);
        _result.push(state.result);
        skipSeparationSpace(state, true, -1);
        ch = state.input.charCodeAt(state.position);
        if ((state.line === _line || state.lineIndent > nodeIndent) && ch !== 0) {
          throwError(state, "bad indentation of a sequence entry");
        } else if (state.lineIndent < nodeIndent) {
          break;
        }
      }
      if (detected) {
        state.tag = _tag2;
        state.anchor = _anchor;
        state.kind = "sequence";
        state.result = _result;
        return true;
      }
      return false;
    }
    function readBlockMapping(state, nodeIndent, flowIndent) {
      let allowCompact;
      let _keyLine;
      let _keyLineStart;
      let _keyPos;
      const _tag2 = state.tag;
      const _anchor = state.anchor;
      const _result = {};
      const overridableKeys = /* @__PURE__ */ Object.create(null);
      let keyTag = null;
      let keyNode = null;
      let valueNode = null;
      let atExplicitKey = false;
      let detected = false;
      if (state.firstTabInLine !== -1) return false;
      if (state.anchor !== null) {
        storeAnchor(state, state.anchor, _result);
      }
      let ch = state.input.charCodeAt(state.position);
      while (ch !== 0) {
        if (!atExplicitKey && state.firstTabInLine !== -1) {
          state.position = state.firstTabInLine;
          throwError(state, "tab characters must not be used in indentation");
        }
        const following = state.input.charCodeAt(state.position + 1);
        const _line = state.line;
        if ((ch === 63 || ch === 58) && isWsOrEol(following)) {
          if (ch === 63) {
            if (atExplicitKey) {
              storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
              keyTag = keyNode = valueNode = null;
            }
            detected = true;
            atExplicitKey = true;
            allowCompact = true;
          } else if (atExplicitKey) {
            atExplicitKey = false;
            allowCompact = true;
          } else {
            throwError(state, "incomplete explicit mapping pair; a key node is missed; or followed by a non-tabulated empty line");
          }
          state.position += 1;
          ch = following;
        } else {
          _keyLine = state.line;
          _keyLineStart = state.lineStart;
          _keyPos = state.position;
          if (!composeNode(state, flowIndent, CONTEXT_FLOW_OUT, false, true)) {
            break;
          }
          if (state.line === _line) {
            ch = state.input.charCodeAt(state.position);
            while (isWhiteSpace(ch)) {
              ch = state.input.charCodeAt(++state.position);
            }
            if (ch === 58) {
              ch = state.input.charCodeAt(++state.position);
              if (!isWsOrEol(ch)) {
                throwError(state, "a whitespace character is expected after the key-value separator within a block mapping");
              }
              if (atExplicitKey) {
                storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
                keyTag = keyNode = valueNode = null;
              }
              detected = true;
              atExplicitKey = false;
              allowCompact = false;
              keyTag = state.tag;
              keyNode = state.result;
            } else if (detected) {
              throwError(state, "can not read an implicit mapping pair; a colon is missed");
            } else {
              state.tag = _tag2;
              state.anchor = _anchor;
              return true;
            }
          } else if (detected) {
            throwError(state, "can not read a block mapping entry; a multiline key may not be an implicit key");
          } else {
            state.tag = _tag2;
            state.anchor = _anchor;
            return true;
          }
        }
        if (state.line === _line || state.lineIndent > nodeIndent) {
          if (atExplicitKey) {
            _keyLine = state.line;
            _keyLineStart = state.lineStart;
            _keyPos = state.position;
          }
          if (composeNode(state, nodeIndent, CONTEXT_BLOCK_OUT, true, allowCompact)) {
            if (atExplicitKey) {
              keyNode = state.result;
            } else {
              valueNode = state.result;
            }
          }
          if (!atExplicitKey) {
            storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, valueNode, _keyLine, _keyLineStart, _keyPos);
            keyTag = keyNode = valueNode = null;
          }
          skipSeparationSpace(state, true, -1);
          ch = state.input.charCodeAt(state.position);
        }
        if ((state.line === _line || state.lineIndent > nodeIndent) && ch !== 0) {
          throwError(state, "bad indentation of a mapping entry");
        } else if (state.lineIndent < nodeIndent) {
          break;
        }
      }
      if (atExplicitKey) {
        storeMappingPair(state, _result, overridableKeys, keyTag, keyNode, null, _keyLine, _keyLineStart, _keyPos);
      }
      if (detected) {
        state.tag = _tag2;
        state.anchor = _anchor;
        state.kind = "mapping";
        state.result = _result;
      }
      return detected;
    }
    function readTagProperty(state) {
      let isVerbatim = false;
      let isNamed = false;
      let tagHandle;
      let tagName;
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 33) return false;
      if (state.tag !== null) {
        throwError(state, "duplication of a tag property");
      }
      ch = state.input.charCodeAt(++state.position);
      if (ch === 60) {
        isVerbatim = true;
        ch = state.input.charCodeAt(++state.position);
      } else if (ch === 33) {
        isNamed = true;
        tagHandle = "!!";
        ch = state.input.charCodeAt(++state.position);
      } else {
        tagHandle = "!";
      }
      let _position = state.position;
      if (isVerbatim) {
        do {
          ch = state.input.charCodeAt(++state.position);
        } while (ch !== 0 && ch !== 62);
        if (state.position < state.length) {
          tagName = state.input.slice(_position, state.position);
          ch = state.input.charCodeAt(++state.position);
        } else {
          throwError(state, "unexpected end of the stream within a verbatim tag");
        }
      } else {
        while (ch !== 0 && !isWsOrEol(ch)) {
          if (ch === 33) {
            if (!isNamed) {
              tagHandle = state.input.slice(_position - 1, state.position + 1);
              if (!PATTERN_TAG_HANDLE.test(tagHandle)) {
                throwError(state, "named tag handle cannot contain such characters");
              }
              isNamed = true;
              _position = state.position + 1;
            } else {
              throwError(state, "tag suffix cannot contain exclamation marks");
            }
          }
          ch = state.input.charCodeAt(++state.position);
        }
        tagName = state.input.slice(_position, state.position);
        if (PATTERN_FLOW_INDICATORS.test(tagName)) {
          throwError(state, "tag suffix cannot contain flow indicator characters");
        }
      }
      if (tagName && !PATTERN_TAG_URI.test(tagName)) {
        throwError(state, "tag name cannot contain such characters: " + tagName);
      }
      try {
        tagName = decodeURIComponent(tagName);
      } catch (err) {
        throwError(state, "tag name is malformed: " + tagName);
      }
      if (isVerbatim) {
        state.tag = tagName;
      } else if (_hasOwnProperty2.call(state.tagMap, tagHandle)) {
        state.tag = state.tagMap[tagHandle] + tagName;
      } else if (tagHandle === "!") {
        state.tag = "!" + tagName;
      } else if (tagHandle === "!!") {
        state.tag = "tag:yaml.org,2002:" + tagName;
      } else {
        throwError(state, 'undeclared tag handle "' + tagHandle + '"');
      }
      return true;
    }
    function readAnchorProperty(state) {
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 38) return false;
      if (state.anchor !== null) {
        throwError(state, "duplication of an anchor property");
      }
      ch = state.input.charCodeAt(++state.position);
      const _position = state.position;
      while (ch !== 0 && !isWsOrEol(ch) && !isFlowIndicator(ch)) {
        ch = state.input.charCodeAt(++state.position);
      }
      if (state.position === _position) {
        throwError(state, "name of an anchor node must contain at least one character");
      }
      state.anchor = state.input.slice(_position, state.position);
      return true;
    }
    function readAlias(state) {
      let ch = state.input.charCodeAt(state.position);
      if (ch !== 42) return false;
      ch = state.input.charCodeAt(++state.position);
      const _position = state.position;
      while (ch !== 0 && !isWsOrEol(ch) && !isFlowIndicator(ch)) {
        ch = state.input.charCodeAt(++state.position);
      }
      if (state.position === _position) {
        throwError(state, "name of an alias node must contain at least one character");
      }
      const alias = state.input.slice(_position, state.position);
      if (!_hasOwnProperty2.call(state.anchorMap, alias)) {
        throwError(state, 'unidentified alias "' + alias + '"');
      }
      state.result = state.anchorMap[alias];
      skipSeparationSpace(state, true, -1);
      return true;
    }
    function tryReadBlockMappingFromProperty(state, propertyStart, nodeIndent, flowIndent) {
      const fallbackState = snapshotState(state);
      beginAnchorTransaction(state);
      restoreState(state, propertyStart);
      state.tag = null;
      state.anchor = null;
      state.kind = null;
      state.result = null;
      if (readBlockMapping(state, nodeIndent, flowIndent) && state.kind === "mapping") {
        commitAnchorTransaction(state);
        return true;
      }
      rollbackAnchorTransaction(state);
      restoreState(state, fallbackState);
      return false;
    }
    function composeNode(state, parentIndent, nodeContext, allowToSeek, allowCompact) {
      let allowBlockScalars;
      let allowBlockCollections;
      let indentStatus = 1;
      let atNewLine = false;
      let hasContent = false;
      let propertyStart = null;
      let type2;
      let flowIndent;
      let blockIndent;
      if (state.depth >= state.maxDepth) {
        throwError(state, "nesting exceeded maxDepth (" + state.maxDepth + ")");
      }
      state.depth += 1;
      if (state.listener !== null) {
        state.listener("open", state);
      }
      state.tag = null;
      state.anchor = null;
      state.kind = null;
      state.result = null;
      const allowBlockStyles = allowBlockScalars = allowBlockCollections = CONTEXT_BLOCK_OUT === nodeContext || CONTEXT_BLOCK_IN === nodeContext;
      if (allowToSeek) {
        if (skipSeparationSpace(state, true, -1)) {
          atNewLine = true;
          if (state.lineIndent > parentIndent) {
            indentStatus = 1;
          } else if (state.lineIndent === parentIndent) {
            indentStatus = 0;
          } else if (state.lineIndent < parentIndent) {
            indentStatus = -1;
          }
        }
      }
      if (indentStatus === 1) {
        while (true) {
          const ch = state.input.charCodeAt(state.position);
          const propertyState = snapshotState(state);
          if (atNewLine && (ch === 33 && state.tag !== null || ch === 38 && state.anchor !== null)) {
            break;
          }
          if (!readTagProperty(state) && !readAnchorProperty(state)) {
            break;
          }
          if (propertyStart === null) {
            propertyStart = propertyState;
          }
          if (skipSeparationSpace(state, true, -1)) {
            atNewLine = true;
            allowBlockCollections = allowBlockStyles;
            if (state.lineIndent > parentIndent) {
              indentStatus = 1;
            } else if (state.lineIndent === parentIndent) {
              indentStatus = 0;
            } else if (state.lineIndent < parentIndent) {
              indentStatus = -1;
            }
          } else {
            allowBlockCollections = false;
          }
        }
      }
      if (allowBlockCollections) {
        allowBlockCollections = atNewLine || allowCompact;
      }
      if (indentStatus === 1 || CONTEXT_BLOCK_OUT === nodeContext) {
        if (CONTEXT_FLOW_IN === nodeContext || CONTEXT_FLOW_OUT === nodeContext) {
          flowIndent = parentIndent;
        } else {
          flowIndent = parentIndent + 1;
        }
        blockIndent = state.position - state.lineStart;
        if (indentStatus === 1) {
          if (allowBlockCollections && (readBlockSequence(state, blockIndent) || readBlockMapping(state, blockIndent, flowIndent)) || readFlowCollection(state, flowIndent)) {
            hasContent = true;
          } else {
            const ch = state.input.charCodeAt(state.position);
            if (propertyStart !== null && allowBlockStyles && !allowBlockCollections && ch !== 124 && ch !== 62 && tryReadBlockMappingFromProperty(
              state,
              propertyStart,
              propertyStart.position - propertyStart.lineStart,
              flowIndent
            )) {
              hasContent = true;
            } else if (allowBlockScalars && readBlockScalar(state, flowIndent) || readSingleQuotedScalar(state, flowIndent) || readDoubleQuotedScalar(state, flowIndent)) {
              hasContent = true;
            } else if (readAlias(state)) {
              hasContent = true;
              if (state.tag !== null || state.anchor !== null) {
                throwError(state, "alias node should not have any properties");
              }
            } else if (readPlainScalar(state, flowIndent, CONTEXT_FLOW_IN === nodeContext)) {
              hasContent = true;
              if (state.tag === null) {
                state.tag = "?";
              }
            }
            if (state.anchor !== null) {
              storeAnchor(state, state.anchor, state.result);
            }
          }
        } else if (indentStatus === 0) {
          hasContent = allowBlockCollections && readBlockSequence(state, blockIndent);
        }
      }
      if (state.tag === null) {
        if (state.anchor !== null) {
          storeAnchor(state, state.anchor, state.result);
        }
      } else if (state.tag === "?") {
        if (state.result !== null && state.kind !== "scalar") {
          throwError(state, 'unacceptable node kind for !<?> tag; it should be "scalar", not "' + state.kind + '"');
        }
        for (let typeIndex = 0, typeQuantity = state.implicitTypes.length; typeIndex < typeQuantity; typeIndex += 1) {
          type2 = state.implicitTypes[typeIndex];
          if (type2.resolve(state.result)) {
            state.result = type2.construct(state.result);
            state.tag = type2.tag;
            if (state.anchor !== null) {
              storeAnchor(state, state.anchor, state.result);
            }
            break;
          }
        }
      } else if (state.tag !== "!") {
        if (_hasOwnProperty2.call(state.typeMap[state.kind || "fallback"], state.tag)) {
          type2 = state.typeMap[state.kind || "fallback"][state.tag];
        } else {
          type2 = null;
          const typeList = state.typeMap.multi[state.kind || "fallback"];
          for (let typeIndex = 0, typeQuantity = typeList.length; typeIndex < typeQuantity; typeIndex += 1) {
            if (state.tag.slice(0, typeList[typeIndex].tag.length) === typeList[typeIndex].tag) {
              type2 = typeList[typeIndex];
              break;
            }
          }
        }
        if (!type2) {
          throwError(state, "unknown tag !<" + state.tag + ">");
        }
        if (state.result !== null && type2.kind !== state.kind) {
          throwError(state, "unacceptable node kind for !<" + state.tag + '> tag; it should be "' + type2.kind + '", not "' + state.kind + '"');
        }
        if (!type2.resolve(state.result, state.tag)) {
          throwError(state, "cannot resolve a node with !<" + state.tag + "> explicit tag");
        } else {
          state.result = type2.construct(state.result, state.tag);
          if (state.anchor !== null) {
            storeAnchor(state, state.anchor, state.result);
          }
        }
      }
      if (state.listener !== null) {
        state.listener("close", state);
      }
      state.depth -= 1;
      return state.tag !== null || state.anchor !== null || hasContent;
    }
    function readDocument(state) {
      const documentStart = state.position;
      let hasDirectives = false;
      let ch;
      state.version = null;
      state.checkLineBreaks = state.legacy;
      state.tagMap = /* @__PURE__ */ Object.create(null);
      state.anchorMap = /* @__PURE__ */ Object.create(null);
      while ((ch = state.input.charCodeAt(state.position)) !== 0) {
        skipSeparationSpace(state, true, -1);
        ch = state.input.charCodeAt(state.position);
        if (state.lineIndent > 0 || ch !== 37) {
          break;
        }
        hasDirectives = true;
        ch = state.input.charCodeAt(++state.position);
        let _position = state.position;
        while (ch !== 0 && !isWsOrEol(ch)) {
          ch = state.input.charCodeAt(++state.position);
        }
        const directiveName = state.input.slice(_position, state.position);
        const directiveArgs = [];
        if (directiveName.length < 1) {
          throwError(state, "directive name must not be less than one character in length");
        }
        while (ch !== 0) {
          while (isWhiteSpace(ch)) {
            ch = state.input.charCodeAt(++state.position);
          }
          if (ch === 35) {
            do {
              ch = state.input.charCodeAt(++state.position);
            } while (ch !== 0 && !isEol(ch));
            break;
          }
          if (isEol(ch)) break;
          _position = state.position;
          while (ch !== 0 && !isWsOrEol(ch)) {
            ch = state.input.charCodeAt(++state.position);
          }
          directiveArgs.push(state.input.slice(_position, state.position));
        }
        if (ch !== 0) readLineBreak(state);
        if (_hasOwnProperty2.call(directiveHandlers, directiveName)) {
          directiveHandlers[directiveName](state, directiveName, directiveArgs);
        } else {
          throwWarning(state, 'unknown document directive "' + directiveName + '"');
        }
      }
      skipSeparationSpace(state, true, -1);
      if (state.lineIndent === 0 && state.input.charCodeAt(state.position) === 45 && state.input.charCodeAt(state.position + 1) === 45 && state.input.charCodeAt(state.position + 2) === 45) {
        state.position += 3;
        skipSeparationSpace(state, true, -1);
      } else if (hasDirectives) {
        throwError(state, "directives end mark is expected");
      }
      composeNode(state, state.lineIndent - 1, CONTEXT_BLOCK_OUT, false, true);
      skipSeparationSpace(state, true, -1);
      if (state.checkLineBreaks && PATTERN_NON_ASCII_LINE_BREAKS.test(state.input.slice(documentStart, state.position))) {
        throwWarning(state, "non-ASCII line breaks are interpreted as content");
      }
      state.documents.push(state.result);
      if (state.position === state.lineStart && testDocumentSeparator(state)) {
        if (state.input.charCodeAt(state.position) === 46) {
          state.position += 3;
          skipSeparationSpace(state, true, -1);
        }
        return;
      }
      if (state.position < state.length - 1) {
        throwError(state, "end of the stream or a document separator is expected");
      }
    }
    function loadDocuments(input, options2) {
      input = String(input);
      options2 = options2 || {};
      if (input.length !== 0) {
        if (input.charCodeAt(input.length - 1) !== 10 && input.charCodeAt(input.length - 1) !== 13) {
          input += "\n";
        }
        if (input.charCodeAt(0) === 65279) {
          input = input.slice(1);
        }
      }
      const state = new State(input, options2);
      const nullpos = input.indexOf("\0");
      if (nullpos !== -1) {
        state.position = nullpos;
        throwError(state, "null byte is not allowed in input");
      }
      state.input += "\0";
      while (state.input.charCodeAt(state.position) === 32) {
        state.lineIndent += 1;
        state.position += 1;
      }
      while (state.position < state.length - 1) {
        readDocument(state);
      }
      return state.documents;
    }
    function loadAll2(input, iterator, options2) {
      if (iterator !== null && typeof iterator === "object" && typeof options2 === "undefined") {
        options2 = iterator;
        iterator = null;
      }
      const documents = loadDocuments(input, options2);
      if (typeof iterator !== "function") {
        return documents;
      }
      for (let index = 0, length = documents.length; index < length; index += 1) {
        iterator(documents[index]);
      }
    }
    function load2(input, options2) {
      const documents = loadDocuments(input, options2);
      if (documents.length === 0) {
        return void 0;
      } else if (documents.length === 1) {
        return documents[0];
      }
      throw new YAMLException2("expected a single document in the stream, but found more");
    }
    loader.loadAll = loadAll2;
    loader.load = load2;
    return loader;
  }
  var dumper = {};
  var hasRequiredDumper;
  function requireDumper() {
    if (hasRequiredDumper) return dumper;
    hasRequiredDumper = 1;
    const common2 = requireCommon();
    const YAMLException2 = requireException();
    const DEFAULT_SCHEMA2 = require_default();
    const _toString = Object.prototype.toString;
    const _hasOwnProperty2 = Object.prototype.hasOwnProperty;
    const CHAR_BOM = 65279;
    const CHAR_TAB = 9;
    const CHAR_LINE_FEED = 10;
    const CHAR_CARRIAGE_RETURN = 13;
    const CHAR_SPACE = 32;
    const CHAR_EXCLAMATION = 33;
    const CHAR_DOUBLE_QUOTE = 34;
    const CHAR_SHARP = 35;
    const CHAR_PERCENT = 37;
    const CHAR_AMPERSAND = 38;
    const CHAR_SINGLE_QUOTE = 39;
    const CHAR_ASTERISK = 42;
    const CHAR_COMMA = 44;
    const CHAR_MINUS = 45;
    const CHAR_COLON = 58;
    const CHAR_EQUALS = 61;
    const CHAR_GREATER_THAN = 62;
    const CHAR_QUESTION = 63;
    const CHAR_COMMERCIAL_AT = 64;
    const CHAR_LEFT_SQUARE_BRACKET = 91;
    const CHAR_RIGHT_SQUARE_BRACKET = 93;
    const CHAR_GRAVE_ACCENT = 96;
    const CHAR_LEFT_CURLY_BRACKET = 123;
    const CHAR_VERTICAL_LINE = 124;
    const CHAR_RIGHT_CURLY_BRACKET = 125;
    const ESCAPE_SEQUENCES = {};
    ESCAPE_SEQUENCES[0] = "\\0";
    ESCAPE_SEQUENCES[7] = "\\a";
    ESCAPE_SEQUENCES[8] = "\\b";
    ESCAPE_SEQUENCES[9] = "\\t";
    ESCAPE_SEQUENCES[10] = "\\n";
    ESCAPE_SEQUENCES[11] = "\\v";
    ESCAPE_SEQUENCES[12] = "\\f";
    ESCAPE_SEQUENCES[13] = "\\r";
    ESCAPE_SEQUENCES[27] = "\\e";
    ESCAPE_SEQUENCES[34] = '\\"';
    ESCAPE_SEQUENCES[92] = "\\\\";
    ESCAPE_SEQUENCES[133] = "\\N";
    ESCAPE_SEQUENCES[160] = "\\_";
    ESCAPE_SEQUENCES[8232] = "\\L";
    ESCAPE_SEQUENCES[8233] = "\\P";
    const DEPRECATED_BOOLEANS_SYNTAX = [
      "y",
      "Y",
      "yes",
      "Yes",
      "YES",
      "on",
      "On",
      "ON",
      "n",
      "N",
      "no",
      "No",
      "NO",
      "off",
      "Off",
      "OFF"
    ];
    const DEPRECATED_BASE60_SYNTAX = /^[-+]?[0-9_]+(?::[0-9_]+)+(?:\.[0-9_]*)?$/;
    function compileStyleMap(schema2, map2) {
      if (map2 === null) return {};
      const result = {};
      const keys = Object.keys(map2);
      for (let index = 0, length = keys.length; index < length; index += 1) {
        let tag2 = keys[index];
        let style = String(map2[tag2]);
        if (tag2.slice(0, 2) === "!!") {
          tag2 = "tag:yaml.org,2002:" + tag2.slice(2);
        }
        const type2 = schema2.compiledTypeMap["fallback"][tag2];
        if (type2 && _hasOwnProperty2.call(type2.styleAliases, style)) {
          style = type2.styleAliases[style];
        }
        result[tag2] = style;
      }
      return result;
    }
    function encodeHex(character) {
      let handle;
      let length;
      const string = character.toString(16).toUpperCase();
      if (character <= 255) {
        handle = "x";
        length = 2;
      } else if (character <= 65535) {
        handle = "u";
        length = 4;
      } else if (character <= 4294967295) {
        handle = "U";
        length = 8;
      } else {
        throw new YAMLException2("code point within a string may not be greater than 0xFFFFFFFF");
      }
      return "\\" + handle + common2.repeat("0", length - string.length) + string;
    }
    const QUOTING_TYPE_SINGLE = 1;
    const QUOTING_TYPE_DOUBLE = 2;
    function State(options2) {
      this.schema = options2["schema"] || DEFAULT_SCHEMA2;
      this.indent = Math.max(1, options2["indent"] || 2);
      this.noArrayIndent = options2["noArrayIndent"] || false;
      this.skipInvalid = options2["skipInvalid"] || false;
      this.flowLevel = common2.isNothing(options2["flowLevel"]) ? -1 : options2["flowLevel"];
      this.styleMap = compileStyleMap(this.schema, options2["styles"] || null);
      this.sortKeys = options2["sortKeys"] || false;
      this.lineWidth = options2["lineWidth"] || 80;
      this.noRefs = options2["noRefs"] || false;
      this.noCompatMode = options2["noCompatMode"] || false;
      this.condenseFlow = options2["condenseFlow"] || false;
      this.quotingType = options2["quotingType"] === '"' ? QUOTING_TYPE_DOUBLE : QUOTING_TYPE_SINGLE;
      this.forceQuotes = options2["forceQuotes"] || false;
      this.replacer = typeof options2["replacer"] === "function" ? options2["replacer"] : null;
      this.implicitTypes = this.schema.compiledImplicit;
      this.explicitTypes = this.schema.compiledExplicit;
      this.tag = null;
      this.result = "";
      this.duplicates = [];
      this.usedDuplicates = null;
    }
    function indentString(string, spaces) {
      const ind = common2.repeat(" ", spaces);
      let position = 0;
      let result = "";
      const length = string.length;
      while (position < length) {
        let line;
        const next = string.indexOf("\n", position);
        if (next === -1) {
          line = string.slice(position);
          position = length;
        } else {
          line = string.slice(position, next + 1);
          position = next + 1;
        }
        if (line.length && line !== "\n") result += ind;
        result += line;
      }
      return result;
    }
    function generateNextLine(state, level) {
      return "\n" + common2.repeat(" ", state.indent * level);
    }
    function testImplicitResolving(state, str2) {
      for (let index = 0, length = state.implicitTypes.length; index < length; index += 1) {
        const type2 = state.implicitTypes[index];
        if (type2.resolve(str2)) {
          return true;
        }
      }
      return false;
    }
    function isWhitespace(c) {
      return c === CHAR_SPACE || c === CHAR_TAB;
    }
    function isPrintable(c) {
      return c >= 32 && c <= 126 || c >= 161 && c <= 55295 && c !== 8232 && c !== 8233 || c >= 57344 && c <= 65533 && c !== CHAR_BOM || c >= 65536 && c <= 1114111;
    }
    function isNsCharOrWhitespace(c) {
      return isPrintable(c) && c !== CHAR_BOM && // - b-char
      c !== CHAR_CARRIAGE_RETURN && c !== CHAR_LINE_FEED;
    }
    function isPlainSafe(c, prev, inblock) {
      const cIsNsCharOrWhitespace = isNsCharOrWhitespace(c);
      const cIsNsChar = cIsNsCharOrWhitespace && !isWhitespace(c);
      return (
        // ns-plain-safe
        (inblock ? cIsNsCharOrWhitespace : cIsNsCharOrWhitespace && // - c-flow-indicator
        c !== CHAR_COMMA && c !== CHAR_LEFT_SQUARE_BRACKET && c !== CHAR_RIGHT_SQUARE_BRACKET && c !== CHAR_LEFT_CURLY_BRACKET && c !== CHAR_RIGHT_CURLY_BRACKET) && // ns-plain-char
        c !== CHAR_SHARP && // false on '#'
        !(prev === CHAR_COLON && !cIsNsChar) || // false on ': '
        isNsCharOrWhitespace(prev) && !isWhitespace(prev) && c === CHAR_SHARP || // change to true on '[^ ]#'
        prev === CHAR_COLON && cIsNsChar
      );
    }
    function isPlainSafeFirst(c) {
      return isPrintable(c) && c !== CHAR_BOM && !isWhitespace(c) && // - s-white
      // - (c-indicator ::=
      // “-” | “?” | “:” | “,” | “[” | “]” | “{” | “}”
      c !== CHAR_MINUS && c !== CHAR_QUESTION && c !== CHAR_COLON && c !== CHAR_COMMA && c !== CHAR_LEFT_SQUARE_BRACKET && c !== CHAR_RIGHT_SQUARE_BRACKET && c !== CHAR_LEFT_CURLY_BRACKET && c !== CHAR_RIGHT_CURLY_BRACKET && // | “#” | “&” | “*” | “!” | “|” | “=” | “>” | “'” | “"”
      c !== CHAR_SHARP && c !== CHAR_AMPERSAND && c !== CHAR_ASTERISK && c !== CHAR_EXCLAMATION && c !== CHAR_VERTICAL_LINE && c !== CHAR_EQUALS && c !== CHAR_GREATER_THAN && c !== CHAR_SINGLE_QUOTE && c !== CHAR_DOUBLE_QUOTE && // | “%” | “@” | “`”)
      c !== CHAR_PERCENT && c !== CHAR_COMMERCIAL_AT && c !== CHAR_GRAVE_ACCENT;
    }
    function isPlainSafeLast(c) {
      return !isWhitespace(c) && c !== CHAR_COLON;
    }
    function codePointAt(string, pos) {
      const first = string.charCodeAt(pos);
      let second;
      if (first >= 55296 && first <= 56319 && pos + 1 < string.length) {
        second = string.charCodeAt(pos + 1);
        if (second >= 56320 && second <= 57343) {
          return (first - 55296) * 1024 + second - 56320 + 65536;
        }
      }
      return first;
    }
    function needIndentIndicator(string) {
      const leadingSpaceRe = /^\n* /;
      return leadingSpaceRe.test(string);
    }
    const STYLE_PLAIN = 1;
    const STYLE_SINGLE = 2;
    const STYLE_LITERAL = 3;
    const STYLE_FOLDED = 4;
    const STYLE_DOUBLE = 5;
    function chooseScalarStyle(string, singleLineOnly, indentPerLevel, lineWidth, testAmbiguousType, quotingType, forceQuotes, inblock) {
      let i;
      let char = 0;
      let prevChar = null;
      let hasLineBreak = false;
      let hasFoldableLine = false;
      const shouldTrackWidth = lineWidth !== -1;
      let previousLineBreak = -1;
      let plain = isPlainSafeFirst(codePointAt(string, 0)) && isPlainSafeLast(codePointAt(string, string.length - 1));
      if (singleLineOnly || forceQuotes) {
        for (i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
          char = codePointAt(string, i);
          if (!isPrintable(char)) {
            return STYLE_DOUBLE;
          }
          plain = plain && isPlainSafe(char, prevChar, inblock);
          prevChar = char;
        }
      } else {
        for (i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
          char = codePointAt(string, i);
          if (char === CHAR_LINE_FEED) {
            hasLineBreak = true;
            if (shouldTrackWidth) {
              hasFoldableLine = hasFoldableLine || // Foldable line = too long, and not more-indented.
              i - previousLineBreak - 1 > lineWidth && string[previousLineBreak + 1] !== " ";
              previousLineBreak = i;
            }
          } else if (!isPrintable(char)) {
            return STYLE_DOUBLE;
          }
          plain = plain && isPlainSafe(char, prevChar, inblock);
          prevChar = char;
        }
        hasFoldableLine = hasFoldableLine || shouldTrackWidth && (i - previousLineBreak - 1 > lineWidth && string[previousLineBreak + 1] !== " ");
      }
      if (!hasLineBreak && !hasFoldableLine) {
        if (plain && !forceQuotes && !testAmbiguousType(string)) {
          return STYLE_PLAIN;
        }
        return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
      }
      if (indentPerLevel > 9 && needIndentIndicator(string)) {
        return STYLE_DOUBLE;
      }
      if (!forceQuotes) {
        return hasFoldableLine ? STYLE_FOLDED : STYLE_LITERAL;
      }
      return quotingType === QUOTING_TYPE_DOUBLE ? STYLE_DOUBLE : STYLE_SINGLE;
    }
    function writeScalar(state, string, level, iskey, inblock) {
      state.dump = (function() {
        if (string.length === 0) {
          return state.quotingType === QUOTING_TYPE_DOUBLE ? '""' : "''";
        }
        if (!state.noCompatMode) {
          if (DEPRECATED_BOOLEANS_SYNTAX.indexOf(string) !== -1 || DEPRECATED_BASE60_SYNTAX.test(string)) {
            return state.quotingType === QUOTING_TYPE_DOUBLE ? '"' + string + '"' : "'" + string + "'";
          }
        }
        const indent = state.indent * Math.max(1, level);
        const lineWidth = state.lineWidth === -1 ? -1 : Math.max(Math.min(state.lineWidth, 40), state.lineWidth - indent);
        const singleLineOnly = iskey || // No block styles in flow mode.
        state.flowLevel > -1 && level >= state.flowLevel;
        function testAmbiguity(string2) {
          return testImplicitResolving(state, string2);
        }
        switch (chooseScalarStyle(
          string,
          singleLineOnly,
          state.indent,
          lineWidth,
          testAmbiguity,
          state.quotingType,
          state.forceQuotes && !iskey,
          inblock
        )) {
          case STYLE_PLAIN:
            return string;
          case STYLE_SINGLE:
            return "'" + string.replace(/'/g, "''") + "'";
          case STYLE_LITERAL:
            return "|" + blockHeader(string, state.indent) + dropEndingNewline(indentString(string, indent));
          case STYLE_FOLDED:
            return ">" + blockHeader(string, state.indent) + dropEndingNewline(indentString(foldString(string, lineWidth), indent));
          case STYLE_DOUBLE:
            return '"' + escapeString(string) + '"';
          default:
            throw new YAMLException2("impossible error: invalid scalar style");
        }
      })();
    }
    function blockHeader(string, indentPerLevel) {
      const indentIndicator = needIndentIndicator(string) ? String(indentPerLevel) : "";
      const clip = string[string.length - 1] === "\n";
      const keep = clip && (string[string.length - 2] === "\n" || string === "\n");
      const chomp = keep ? "+" : clip ? "" : "-";
      return indentIndicator + chomp + "\n";
    }
    function dropEndingNewline(string) {
      return string[string.length - 1] === "\n" ? string.slice(0, -1) : string;
    }
    function foldString(string, width) {
      const lineRe = /(\n+)([^\n]*)/g;
      let result = (function() {
        let nextLF = string.indexOf("\n");
        nextLF = nextLF !== -1 ? nextLF : string.length;
        lineRe.lastIndex = nextLF;
        return foldLine(string.slice(0, nextLF), width);
      })();
      let prevMoreIndented = string[0] === "\n" || string[0] === " ";
      let moreIndented;
      let match;
      while (match = lineRe.exec(string)) {
        const prefix = match[1];
        const line = match[2];
        moreIndented = line[0] === " ";
        result += prefix + (!prevMoreIndented && !moreIndented && line !== "" ? "\n" : "") + foldLine(line, width);
        prevMoreIndented = moreIndented;
      }
      return result;
    }
    function foldLine(line, width) {
      if (line === "" || line[0] === " ") return line;
      const breakRe = / [^ ]/g;
      let match;
      let start = 0;
      let end;
      let curr = 0;
      let next = 0;
      let result = "";
      while (match = breakRe.exec(line)) {
        next = match.index;
        if (next - start > width) {
          end = curr > start ? curr : next;
          result += "\n" + line.slice(start, end);
          start = end + 1;
        }
        curr = next;
      }
      result += "\n";
      if (line.length - start > width && curr > start) {
        result += line.slice(start, curr) + "\n" + line.slice(curr + 1);
      } else {
        result += line.slice(start);
      }
      return result.slice(1);
    }
    function escapeString(string) {
      let result = "";
      let char = 0;
      for (let i = 0; i < string.length; char >= 65536 ? i += 2 : i++) {
        char = codePointAt(string, i);
        const escapeSeq = ESCAPE_SEQUENCES[char];
        if (!escapeSeq && isPrintable(char)) {
          result += string[i];
          if (char >= 65536) result += string[i + 1];
        } else {
          result += escapeSeq || encodeHex(char);
        }
      }
      return result;
    }
    function writeFlowSequence(state, level, object) {
      let _result = "";
      const _tag2 = state.tag;
      for (let index = 0, length = object.length; index < length; index += 1) {
        let value = object[index];
        if (state.replacer) {
          value = state.replacer.call(object, String(index), value);
        }
        if (writeNode(state, level, value, false, false) || typeof value === "undefined" && writeNode(state, level, null, false, false)) {
          if (_result !== "") _result += "," + (!state.condenseFlow ? " " : "");
          _result += state.dump;
        }
      }
      state.tag = _tag2;
      state.dump = "[" + _result + "]";
    }
    function writeBlockSequence(state, level, object, compact) {
      let _result = "";
      const _tag2 = state.tag;
      for (let index = 0, length = object.length; index < length; index += 1) {
        let value = object[index];
        if (state.replacer) {
          value = state.replacer.call(object, String(index), value);
        }
        if (writeNode(state, level + 1, value, true, true, false, true) || typeof value === "undefined" && writeNode(state, level + 1, null, true, true, false, true)) {
          if (!compact || _result !== "") {
            _result += generateNextLine(state, level);
          }
          if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
            _result += "-";
          } else {
            _result += "- ";
          }
          _result += state.dump;
        }
      }
      state.tag = _tag2;
      state.dump = _result || "[]";
    }
    function writeFlowMapping(state, level, object) {
      let _result = "";
      const _tag2 = state.tag;
      const objectKeyList = Object.keys(object);
      for (let index = 0, length = objectKeyList.length; index < length; index += 1) {
        let pairBuffer = "";
        if (_result !== "") pairBuffer += ", ";
        if (state.condenseFlow) pairBuffer += '"';
        const objectKey = objectKeyList[index];
        let objectValue = object[objectKey];
        if (state.replacer) {
          objectValue = state.replacer.call(object, objectKey, objectValue);
        }
        if (!writeNode(state, level, objectKey, false, false)) {
          continue;
        }
        if (state.dump.length > 1024) pairBuffer += "? ";
        pairBuffer += state.dump + (state.condenseFlow ? '"' : "") + ":" + (state.condenseFlow ? "" : " ");
        if (!writeNode(state, level, objectValue, false, false)) {
          continue;
        }
        pairBuffer += state.dump;
        _result += pairBuffer;
      }
      state.tag = _tag2;
      state.dump = "{" + _result + "}";
    }
    function writeBlockMapping(state, level, object, compact) {
      let _result = "";
      const _tag2 = state.tag;
      const objectKeyList = Object.keys(object);
      if (state.sortKeys === true) {
        objectKeyList.sort();
      } else if (typeof state.sortKeys === "function") {
        objectKeyList.sort(state.sortKeys);
      } else if (state.sortKeys) {
        throw new YAMLException2("sortKeys must be a boolean or a function");
      }
      for (let index = 0, length = objectKeyList.length; index < length; index += 1) {
        let pairBuffer = "";
        if (!compact || _result !== "") {
          pairBuffer += generateNextLine(state, level);
        }
        const objectKey = objectKeyList[index];
        let objectValue = object[objectKey];
        if (state.replacer) {
          objectValue = state.replacer.call(object, objectKey, objectValue);
        }
        if (!writeNode(state, level + 1, objectKey, true, true, true)) {
          continue;
        }
        const explicitPair = state.tag !== null && state.tag !== "?" || state.dump && state.dump.length > 1024;
        if (explicitPair) {
          if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
            pairBuffer += "?";
          } else {
            pairBuffer += "? ";
          }
        }
        pairBuffer += state.dump;
        if (explicitPair) {
          pairBuffer += generateNextLine(state, level);
        }
        if (!writeNode(state, level + 1, objectValue, true, explicitPair)) {
          continue;
        }
        if (state.dump && CHAR_LINE_FEED === state.dump.charCodeAt(0)) {
          pairBuffer += ":";
        } else {
          pairBuffer += ": ";
        }
        pairBuffer += state.dump;
        _result += pairBuffer;
      }
      state.tag = _tag2;
      state.dump = _result || "{}";
    }
    function detectType(state, object, explicit) {
      const typeList = explicit ? state.explicitTypes : state.implicitTypes;
      for (let index = 0, length = typeList.length; index < length; index += 1) {
        const type2 = typeList[index];
        if ((type2.instanceOf || type2.predicate) && (!type2.instanceOf || typeof object === "object" && object instanceof type2.instanceOf) && (!type2.predicate || type2.predicate(object))) {
          if (explicit) {
            if (type2.multi && type2.representName) {
              state.tag = type2.representName(object);
            } else {
              state.tag = type2.tag;
            }
          } else {
            state.tag = "?";
          }
          if (type2.represent) {
            const style = state.styleMap[type2.tag] || type2.defaultStyle;
            let _result;
            if (_toString.call(type2.represent) === "[object Function]") {
              _result = type2.represent(object, style);
            } else if (_hasOwnProperty2.call(type2.represent, style)) {
              _result = type2.represent[style](object, style);
            } else {
              throw new YAMLException2("!<" + type2.tag + '> tag resolver accepts not "' + style + '" style');
            }
            state.dump = _result;
          }
          return true;
        }
      }
      return false;
    }
    function writeNode(state, level, object, block2, compact, iskey, isblockseq) {
      state.tag = null;
      state.dump = object;
      if (!detectType(state, object, false)) {
        detectType(state, object, true);
      }
      const type2 = _toString.call(state.dump);
      const inblock = block2;
      if (block2) {
        block2 = state.flowLevel < 0 || state.flowLevel > level;
      }
      const objectOrArray = type2 === "[object Object]" || type2 === "[object Array]";
      let duplicateIndex;
      let duplicate;
      if (objectOrArray) {
        duplicateIndex = state.duplicates.indexOf(object);
        duplicate = duplicateIndex !== -1;
      }
      if (state.tag !== null && state.tag !== "?" || duplicate || state.indent !== 2 && level > 0) {
        compact = false;
      }
      if (duplicate && state.usedDuplicates[duplicateIndex]) {
        state.dump = "*ref_" + duplicateIndex;
      } else {
        if (objectOrArray && duplicate && !state.usedDuplicates[duplicateIndex]) {
          state.usedDuplicates[duplicateIndex] = true;
        }
        if (type2 === "[object Object]") {
          if (block2 && Object.keys(state.dump).length !== 0) {
            writeBlockMapping(state, level, state.dump, compact);
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + state.dump;
            }
          } else {
            writeFlowMapping(state, level, state.dump);
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + " " + state.dump;
            }
          }
        } else if (type2 === "[object Array]") {
          if (block2 && state.dump.length !== 0) {
            if (state.noArrayIndent && !isblockseq && level > 0) {
              writeBlockSequence(state, level - 1, state.dump, compact);
            } else {
              writeBlockSequence(state, level, state.dump, compact);
            }
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + state.dump;
            }
          } else {
            writeFlowSequence(state, level, state.dump);
            if (duplicate) {
              state.dump = "&ref_" + duplicateIndex + " " + state.dump;
            }
          }
        } else if (type2 === "[object String]") {
          if (state.tag !== "?") {
            writeScalar(state, state.dump, level, iskey, inblock);
          }
        } else if (type2 === "[object Undefined]") {
          return false;
        } else {
          if (state.skipInvalid) return false;
          throw new YAMLException2("unacceptable kind of an object to dump " + type2);
        }
        if (state.tag !== null && state.tag !== "?") {
          let tagStr = encodeURI(
            state.tag[0] === "!" ? state.tag.slice(1) : state.tag
          ).replace(/!/g, "%21");
          if (state.tag[0] === "!") {
            tagStr = "!" + tagStr;
          } else if (tagStr.slice(0, 18) === "tag:yaml.org,2002:") {
            tagStr = "!!" + tagStr.slice(18);
          } else {
            tagStr = "!<" + tagStr + ">";
          }
          state.dump = tagStr + " " + state.dump;
        }
      }
      return true;
    }
    function getDuplicateReferences(object, state) {
      const objects = [];
      const duplicatesIndexes = [];
      inspectNode(object, objects, duplicatesIndexes);
      const length = duplicatesIndexes.length;
      for (let index = 0; index < length; index += 1) {
        state.duplicates.push(objects[duplicatesIndexes[index]]);
      }
      state.usedDuplicates = new Array(length);
    }
    function inspectNode(object, objects, duplicatesIndexes) {
      if (object !== null && typeof object === "object") {
        const index = objects.indexOf(object);
        if (index !== -1) {
          if (duplicatesIndexes.indexOf(index) === -1) {
            duplicatesIndexes.push(index);
          }
        } else {
          objects.push(object);
          if (Array.isArray(object)) {
            for (let i = 0, length = object.length; i < length; i += 1) {
              inspectNode(object[i], objects, duplicatesIndexes);
            }
          } else {
            const objectKeyList = Object.keys(object);
            for (let i = 0, length = objectKeyList.length; i < length; i += 1) {
              inspectNode(object[objectKeyList[i]], objects, duplicatesIndexes);
            }
          }
        }
      }
    }
    function dump2(input, options2) {
      options2 = options2 || {};
      const state = new State(options2);
      if (!state.noRefs) getDuplicateReferences(input, state);
      let value = input;
      if (state.replacer) {
        value = state.replacer.call({ "": value }, "", value);
      }
      if (writeNode(state, 0, value, true, true)) return state.dump + "\n";
      return "";
    }
    dumper.dump = dump2;
    return dumper;
  }
  var hasRequiredJsYaml;
  function requireJsYaml() {
    if (hasRequiredJsYaml) return jsYaml;
    hasRequiredJsYaml = 1;
    const loader2 = requireLoader();
    const dumper2 = requireDumper();
    function renamed(from, to) {
      return function() {
        throw new Error("Function yaml." + from + " is removed in js-yaml 4. Use yaml." + to + " instead, which is now safe by default.");
      };
    }
    jsYaml.Type = requireType();
    jsYaml.Schema = requireSchema();
    jsYaml.FAILSAFE_SCHEMA = requireFailsafe();
    jsYaml.JSON_SCHEMA = requireJson();
    jsYaml.CORE_SCHEMA = requireCore();
    jsYaml.DEFAULT_SCHEMA = require_default();
    jsYaml.load = loader2.load;
    jsYaml.loadAll = loader2.loadAll;
    jsYaml.dump = dumper2.dump;
    jsYaml.YAMLException = requireException();
    jsYaml.types = {
      binary: requireBinary(),
      float: requireFloat(),
      map: requireMap(),
      null: require_null(),
      pairs: requirePairs(),
      set: requireSet(),
      timestamp: requireTimestamp(),
      bool: requireBool(),
      int: requireInt(),
      merge: requireMerge(),
      omap: requireOmap(),
      seq: requireSeq(),
      str: requireStr()
    };
    jsYaml.safeLoad = renamed("safeLoad", "load");
    jsYaml.safeLoadAll = renamed("safeLoadAll", "loadAll");
    jsYaml.safeDump = renamed("safeDump", "dump");
    return jsYaml;
  }
  var jsYamlExports = requireJsYaml();
  var yaml = /* @__PURE__ */ getDefaultExportFromCjs(jsYamlExports);
  var {
    Type,
    Schema,
    FAILSAFE_SCHEMA,
    JSON_SCHEMA,
    CORE_SCHEMA,
    DEFAULT_SCHEMA,
    load,
    loadAll,
    dump,
    YAMLException,
    types,
    safeLoad,
    safeLoadAll,
    safeDump
  } = yaml;

  // src/render/md.js
  marked.setOptions({ gfm: true, breaks: true, async: false });
  var PURIFY_CFG = {
    ALLOWED_TAGS: [
      "a",
      "p",
      "br",
      "hr",
      "h1",
      "h2",
      "h3",
      "h4",
      "h5",
      "h6",
      "ul",
      "ol",
      "li",
      "blockquote",
      "pre",
      "code",
      "em",
      "strong",
      "del",
      "table",
      "thead",
      "tbody",
      "tr",
      "th",
      "td",
      "span",
      "img"
    ],
    ALLOWED_ATTR: ["href", "src", "alt", "title"],
    ALLOW_DATA_ATTR: false
    // URI 白名单沿用 DOMPurify 默认（http(s)/mailto/相对——javascript:/data: 默认已封）。
  };
  function purifySupported() {
    return !!(purify_default && purify_default.isSupported);
  }
  function sanitizeHtml(html3) {
    if (!purifySupported()) return "";
    return purify_default.sanitize(String(html3 == null ? "" : html3), PURIFY_CFG);
  }
  function inline2(s) {
    let out = String(s).replace(/`([^`]+)`/g, '<code class="md-code">$1</code>').replace(/\*\*([^*]+)\*\*/g, "<strong>$1</strong>").replace(/(^|[^*])\*([^*\n]+)\*/g, "$1<em>$2</em>");
    out = out.replace(/\[([^\]]+)\]\(([^)\s]+)\)/g, (m, text2, href) => {
      const ok = safeHref(href);
      return ok ? '<a class="md-a" href="' + esc(ok) + '">' + text2 + "</a>" : text2 + ' <span class="mono faint small">(' + href + ")</span>";
    });
    return out;
  }
  function minimalRender(src) {
    const lines = String(src == null ? "" : src).split(/\r?\n/);
    const out = [];
    const para = [];
    const flush = () => {
      if (para.length) {
        out.push('<p class="md-p">' + inline2(esc(para.join(" "))) + "</p>");
        para.length = 0;
      }
    };
    const isList = (l) => /^(\s*)([-*+]|\d+[.)])\s+/.test(l);
    let i = 0;
    while (i < lines.length) {
      const line = lines[i];
      if (!line.trim()) {
        flush();
        i += 1;
        continue;
      }
      if (/^\s*```/.test(line)) {
        flush();
        const buf = [];
        i += 1;
        while (i < lines.length && !/^\s*```/.test(lines[i])) {
          buf.push(lines[i]);
          i += 1;
        }
        i += 1;
        out.push('<pre class="md-pre">' + esc(buf.join("\n")) + "</pre>");
        continue;
      }
      const h = /^(#{1,6})\s+(.*)$/.exec(line);
      if (h) {
        flush();
        const lv = Math.min(h[1].length, 6);
        out.push("<h" + lv + ' class="md-h md-h' + lv + '">' + inline2(esc(h[2])) + "</h" + lv + ">");
        i += 1;
        continue;
      }
      if (/^\s*(-{3,}|\*{3,}|_{3,})\s*$/.test(line)) {
        flush();
        out.push('<hr class="md-hr">');
        i += 1;
        continue;
      }
      if (/^\s*>\s?/.test(line)) {
        flush();
        const q = [];
        while (i < lines.length && /^\s*>\s?/.test(lines[i])) {
          q.push(lines[i].replace(/^\s*>\s?/, ""));
          i += 1;
        }
        out.push('<blockquote class="md-quote">' + inline2(esc(q.join(" "))) + "</blockquote>");
        continue;
      }
      if (/^\s*\|.+\|\s*$/.test(line) && i + 1 < lines.length && /^\s*\|[\s:|-]+\|\s*$/.test(lines[i + 1])) {
        flush();
        const heads = line.trim().replace(/^\||\|$/g, "").split("|").map((x) => x.trim());
        i += 2;
        const body = [];
        while (i < lines.length && /^\s*\|.+\|\s*$/.test(lines[i])) {
          body.push(lines[i].trim().replace(/^\||\|$/g, "").split("|").map((x) => x.trim()));
          i += 1;
        }
        out.push('<table class="md-tbl"><thead><tr>' + heads.map((x) => "<th>" + inline2(esc(x)) + "</th>").join("") + "</tr></thead><tbody>" + body.map((row) => "<tr>" + heads.map((_, ci) => "<td>" + inline2(esc(row[ci] == null ? "" : row[ci])) + "</td>").join("") + "</tr>").join("") + "</tbody></table>");
        continue;
      }
      if (isList(line)) {
        flush();
        const items = [];
        while (i < lines.length && isList(lines[i])) {
          const m = /^(\s*)([-*+]|\d+[.)])\s+(.*)$/.exec(lines[i]);
          items.push({ depth: m[1].length >= 2 ? 1 : 0, ord: /\d/.test(m[2]), text: m[3] });
          i += 1;
        }
        let html3 = "<" + (items[0].ord ? "ol" : "ul") + ' class="md-list md-list-root">';
        const stack = [{ tag: items[0].ord ? "ol" : "ul" }];
        let openLi = false;
        for (const it of items) {
          if (openLi) {
            html3 += "</li>";
            openLi = false;
          }
          const want = it.ord ? "ol" : "ul";
          while (stack.length > it.depth + 1) html3 += "</" + stack.pop().tag + ">";
          if (stack.length === it.depth + 1 && stack[stack.length - 1].tag !== want) {
            html3 += "</" + stack.pop().tag + "><" + want + ' class="md-list">';
            stack.push({ tag: want });
          }
          while (stack.length < it.depth + 1) {
            html3 += "<" + want + ' class="md-list">';
            stack.push({ tag: want });
          }
          html3 += '<li class="md-li">' + inline2(esc(it.text));
          openLi = true;
        }
        if (openLi) html3 += "</li>";
        while (stack.length) html3 += "</" + stack.pop().tag + ">";
        out.push('<div class="md-wrap">' + html3 + "</div>");
        continue;
      }
      para.push(line.trim());
      i += 1;
    }
    flush();
    return out.join("");
  }
  function renderMarkdown(src) {
    const s = String(src == null ? "" : src);
    if (purifySupported()) {
      let raw = "";
      try {
        raw = marked.parse(s);
      } catch {
        raw = "";
      }
      const clean2 = sanitizeHtml(raw);
      return clean2 || minimalRender(s);
    }
    return minimalRender(s);
  }
  function loadYaml(src) {
    try {
      const v = yaml.load(String(src == null ? "" : src), { json: true });
      return v && typeof v === "object" ? v : null;
    } catch {
      return null;
    }
  }
  var FIELD_GLOSS = {
    name: t("\u540D\u79F0"),
    scope: t("\u4F5C\u7528\u57DF"),
    resourceVersion: t("\u7248\u672C\u53F7"),
    goalRef: t("\u76EE\u6807\u6587\u6863"),
    descRef: t("\u8BF4\u660E\u6587\u6863"),
    docs: t("\u76F8\u5173\u6587\u6863"),
    externalRefs: t("\u5916\u90E8\u4F9D\u8D56"),
    instances: t("\u670D\u52A1\u5B9E\u4F8B"),
    builtinServices: t("\u5185\u7F6E\u670D\u52A1"),
    agentRefs: t("\u5BF9\u8BDD\u5165\u53E3\uFF08Agent\uFF09"),
    phase: t("\u8FD0\u884C\u72B6\u6001"),
    observedGeneration: t("\u5DF2\u5E94\u7528\u7684\u7248\u672C"),
    labels: t("\u6807\u7B7E"),
    mounts: t("\u5DF2\u6302\u8F7D\u8FDE\u63A5"),
    connectorOps: t("\u5DF2\u6302\u8F7D\u8FDE\u63A5"),
    provider: t("\u51ED\u636E\u6E90"),
    session: t("\u5F53\u524D\u4F1A\u8BDD"),
    tags: t("\u6807\u7B7E"),
    title: t("\u6807\u9898"),
    dependsOn: t("\u4F9D\u8D56"),
    agent: t("\u8D1F\u8D23")
  };
  function cardVal(v, depth) {
    if (v == null) return '<span class="faint">\u2014</span>';
    if (Array.isArray(v)) {
      if (!v.length) return t('<span class="faint">\uFF08\u7A7A\uFF09</span>');
      return '<div class="md-chips">' + v.slice(0, 24).map((x) => '<span class="chip mute small md-chip">' + cardVal(x, depth + 1) + "</span>").join("") + (v.length > 24 ? t('<span class="faint small">\u2026\u5171 ') + v.length + t(" \u9879</span>") : "") + "</div>";
    }
    if (typeof v === "object") {
      if (depth >= 4) return t('<span class="mono faint small">\uFF08\u5D4C\u5957\u7ED3\u6784\u89C1\u300CYAML\u300D\uFF09</span>');
      const keys = Object.keys(v);
      if (!keys.length) return t('<span class="faint">\uFF08\u7A7A\uFF09</span>');
      return '<div class="md-subcards">' + keys.map(
        (k) => '<div class="md-subcard"><div class="md-subcard-k">' + esc(k) + '</div><div class="md-subcard-v">' + cardVal(v[k], depth + 1) + "</div></div>"
      ).join("") + "</div>";
    }
    const s = String(v);
    if (/^(agent|system|session):/.test(s)) return '<span class="lk md-lk">' + esc(s) + "</span>";
    if (/^https?:\/\//i.test(s)) {
      const href = safeHref(s);
      return href ? '<a class="md-a" href="' + esc(href) + '">' + esc(trunc(s, 80)) + "</a>" : esc(trunc(s, 80));
    }
    return esc(trunc(s, 300));
  }
  function cardHtml(key, v, depth) {
    const g = FIELD_GLOSS[String(key)];
    return '<div class="md-card"><div class="md-card-k">' + esc(String(key)) + (g ? '<span class="md-gloss">' + esc(g) + "</span>" : "") + '</div><div class="md-card-v">' + cardVal(v, depth) + "</div></div>";
  }
  function yamlToCards(src) {
    const tree = loadYaml(src);
    if (!tree || !Object.keys(tree).length) return "";
    return '<div class="md-cards">' + Object.keys(tree).map((k) => cardHtml(k, tree[k], 0)).join("") + "</div>";
  }
  function rawDisclosure(title, content) {
    return '<details class="md-raw"><summary>' + esc(title || t("\u539F\u6587")) + '</summary><pre class="mono md-raw-pre">' + esc(content == null ? "" : content) + "</pre></details>";
  }

  // src/screens/sessionctx.js
  var PANEL = /* @__PURE__ */ new Map();
  function parseTagInput(raw) {
    return String(raw == null ? "" : raw).split(/[\s,，、]+/).map((t2) => t2.trim()).filter(Boolean);
  }
  function capTags(tags, maxTotal = 20, maxPerType = 5) {
    const kept = [];
    const dropped = [];
    const seen = /* @__PURE__ */ new Set();
    const perType = /* @__PURE__ */ new Map();
    for (const raw of Array.isArray(tags) ? tags : []) {
      const key = String(raw || "").trim();
      if (!key || seen.has(key)) continue;
      seen.add(key);
      const c = key.indexOf(":");
      const typ = c > 0 ? key.slice(0, c) : "";
      const n = perType.get(typ) || 0;
      if (kept.length >= maxTotal || n >= maxPerType) {
        dropped.push(key);
        continue;
      }
      perType.set(typ, n + 1);
      kept.push(key);
    }
    return { kept, dropped };
  }
  function normalizeContext(data) {
    const d = data && typeof data === "object" ? data : {};
    const pack = d.pack && typeof d.pack === "object" ? d.pack : {};
    const pol = d.policy && typeof d.policy === "object" ? d.policy : {};
    const sess = d.session && typeof d.session === "object" ? d.session : null;
    const ov = d.override && typeof d.override === "object" ? d.override : null;
    const chunks = (Array.isArray(pack.chunks) ? pack.chunks : []).map((c) => {
      const row = c && typeof c === "object" ? c : {};
      return {
        kind: String(row.kind || ""),
        scope: String(row.scope || ""),
        name: String(row.name || ""),
        score: Number(row.score) || 0,
        why: String(row.why || "")
      };
    });
    const shards = (Array.isArray(d.shards) ? d.shards : []).map((s) => {
      const row = s && typeof s === "object" ? s : {};
      return {
        tagKey: String(row.tagKey || ""),
        docs: Number(row.docs) || 0,
        topDocs: (Array.isArray(row.topDocs) ? row.topDocs : []).map(String),
        stale: !!row.stale
      };
    });
    const stats = d.shardsStats && typeof d.shardsStats === "object" ? {
      total: Number(d.shardsStats.total) || 0,
      stale: Number(d.shardsStats.stale) || 0,
      threshold: Number(d.shardsStats.rebuildThreshold) || 0
    } : null;
    return {
      present: !!(d.sessionId || d.pack || d.policy),
      sessionId: String(d.sessionId || ""),
      session: sess ? {
        state: String(sess.state || ""),
        phase: String(sess.phase || ""),
        lastAction: String(sess.lastAction || ""),
        heartbeats: Number(sess.heartbeats) || 0,
        lastBeat: String(sess.lastBeat || "")
      } : null,
      takeBudget: Number(pack.takeBudget) || 0,
      takeBudgetDefault: Number(pack.takeBudgetDefault) || 0,
      tags: (Array.isArray(pack.tagSig) ? pack.tagSig : []).map((t2) => String((t2 && typeof t2 === "object" ? t2.tagKey : t2) || "")).filter(Boolean),
      tagSigSource: String(pack.tagSigSource || "none"),
      chunks,
      shards,
      stats,
      shardsFacePresent: Array.isArray(d.shards),
      degraded: pack.degraded !== false,
      degradedReason: String(pack.degradedReason || ""),
      policy: {
        shardMaxTokens: Number(pol.shardMaxTokens) || 0,
        minTakeBudget: Number(pol.minTakeBudget) || 0,
        maxTakeBudget: Number(pol.maxTakeBudget) || 0,
        maxTagsPerType: Number(pol.maxTagsPerType) || 5,
        maxTagsTotal: Number(pol.maxTagsTotal) || 20
      },
      override: ov ? {
        revision: Number(ov.revision) || 0,
        takeBudget: Number(ov.takeBudget) || 0,
        updatedAt: String(ov.updatedAt || ""),
        note: String(ov.note || ""),
        by: String(ov.by || "")
      } : null,
      canRevert: !!d.canRevert,
      writeOpen: !(d.writeGate && d.writeGate.open === false)
    };
  }
  function buildSavePayload(model, input) {
    const pol = model && model.policy || {};
    const min = Number(pol.minTakeBudget) || 0;
    const max = Number(pol.maxTakeBudget) || 0;
    const budget = Number(input && input.takeBudget);
    if (!Number.isFinite(budget) || budget < min || budget > max) {
      return { ok: false, error: t("takeBudget \u987B\u4E3A ") + min + "\u2013" + max + t(" \u7684\u6570\u503C\uFF08ContextPolicy \u9884\u7B97\u57DF\uFF09") };
    }
    const { kept, dropped } = capTags(
      parseTagInput(input && input.tagList),
      Number(pol.maxTagsTotal) || 20,
      Number(pol.maxTagsPerType) || 5
    );
    return {
      ok: true,
      dropped,
      payload: {
        takeBudget: Math.round(budget),
        tagSig: kept,
        note: String((input && input.note) == null ? "" : input.note).slice(0, 512),
        by: "console:web",
        baseRevision: model && model.override ? Number(model.override.revision) || 0 : 0
      }
    };
  }
  function sessionRowHtml(m) {
    if (!m.session) {
      return t('<div class="small faint">\u4F1A\u8BDD\u4E0D\u5728\u6D3B\u6027\u8D26\uFF08\u65E0\u5FC3\u8DF3\u884C\uFF09\u2014\u2014\u5951\u7EA6\u9762\u4ECD\u53EF\u67E5\u770B/\u7F16\u8F91\u672C\u4F1A\u8BDD\u4E0A\u4E0B\u6587\u8986\u76D6\uFF08\u4E0D\u5192\u5145\u5728\u518C\uFF09\u3002</div>');
    }
    const s = m.session;
    const cls = s.state === "suspected-stall" || s.state === "escalated" ? "esc" : s.state ? "acc" : "mute";
    return kv([
      ["state", s.state ? chip(s.state, cls) : '<span class="faint">\u2014</span>'],
      ["phase", esc(s.phase || "\u2014")],
      [t("\u5FC3\u8DF3"), esc(fmtNum(s.heartbeats)) + t(" \u6B21") + (s.lastBeat ? t(" \xB7 \u6700\u8FD1 ") + esc(relAge(s.lastBeat) || t("\u65F6\u95F4\u672A\u77E5")) : "")],
      [t("\u52A8\u4F5C"), esc(trunc(s.lastAction || "\u2014", 80))]
    ]);
  }
  function chunksHtml(m) {
    if (!m.chunks.length) {
      return t('<div class="small faint">\u65E0\u7D22\u5F15\u5019\u9009\uFF08tagSig \u672A\u88C5\u914D\u6216\u7D22\u5F15\u96F6\u547D\u4E2D\u2014\u2014\u4E0D\u5192\u5145\u88C5\u914D\u5305\u5185\u5BB9\uFF09\u3002</div>');
    }
    return m.chunks.map((c) => {
      const key = [c.kind, c.scope, c.name].filter(Boolean).join("/");
      return '<div class="mini-row"><div class="mr-main"><div><b class="mono">' + esc(trunc(key, 64)) + "</b> " + chip(String(Math.round(c.score)), c.score >= 60 ? "ok" : "mute") + "</div>" + (c.why ? '<div class="mr-sub small">' + esc(trunc(c.why, 90)) + "</div>" : "") + "</div></div>";
    }).join("");
  }
  function shardsHtml(m) {
    if (!m.shardsFacePresent) {
      return t('<div class="small faint">\u5206\u7247\u7F13\u5B58\u9762\u7F3A\u5E2D\uFF08contextidx \u672A\u66B4\u9732 Shards\u2014\u2014\u8BDA\u5B9E\u964D\u7EA7\uFF0C\u4E0D\u5192\u5145\u7A7A\u7F13\u5B58\uFF09\u3002</div>');
    }
    if (!m.shards.length) {
      return t('<div class="small faint">\u65E0\u751F\u6548 tag \u5206\u7247\uFF08tagSig \u672A\u88C5\u914D\u2014\u2014\u7F16\u8F91\u5206\u53C9\u540E\u9010 tagKey \u6295\u5F71 TagCtxShard\uFF09\u3002</div>');
    }
    return m.shards.map(
      (s) => '<div class="mini-row"><div class="mr-main"><div><b class="mono">' + esc(s.tagKey) + "</b> " + (s.stale ? chip("stale", "warn") : chip("fresh", "ok")) + "</div>" + t('<div class="mr-sub small">\u76F8\u5173\u5BF9\u8C61 ') + esc(fmtNum(s.docs)) + t(" \u4E2A") + (s.topDocs.length ? " \xB7 " + esc(trunc(s.topDocs.join(", "), 60)) : "") + "</div></div></div>"
    ).join("");
  }
  function formHtml(m) {
    if (!m.writeOpen) {
      return t('<div class="small faint">\u5199\u9762\u5173\u65AD\uFF08--allow-apply \u672A\u5F00\u542F\uFF09\u2014\u2014\u7F16\u8F91\u5206\u53C9\u4E0D\u53EF\u7528\uFF08\u4E0E\u5FC3\u8DF3\u6CE8\u518C\u9762\u540C\u5F00\u5173\uFF0C\u4E0D\u5192\u5145\u53EF\u7F16\u8F91\uFF09\u3002</div>');
    }
    return '<div class="filter-bar" data-ctx-form><input class="inp" data-ctx-budget type="number" min="' + m.policy.minTakeBudget + '" max="' + m.policy.maxTakeBudget + '" step="500" value="' + m.takeBudget + t('" title="\u63D0\u53D6\u9884\u7B97\uFF08token \u6570\uFF09">') + t('<input class="inp" data-ctx-tags placeholder="tag \u96C6\uFF1AtagType:tag \u9017\u53F7/\u7A7A\u683C\u5206\u9694\uFF08\u6BCF\u7C7B\u2264') + m.policy.maxTagsPerType + t(" \u603B\u2264") + m.policy.maxTagsTotal + '\uFF09">' + t('<input class="inp" data-ctx-note placeholder="\u5206\u53C9\u5907\u6CE8\uFF08\u53EF\u9009\uFF09">') + t('<button class="btn primary sm" data-ctx-save>\u4FDD\u5B58\u5206\u53C9</button>') + '<button class="btn sm" data-ctx-revert' + (m.canRevert ? "" : " disabled") + t(">\u64A4\u9500\u4E0A\u6B21</button>") + "</div>";
  }
  function renderSessionCtx(m) {
    const headChips = (m.degraded ? chip(t("\u7D22\u5F15\u5019\u9009\u6295\u5F71"), "warn") : "") + (m.writeOpen ? chip(t("\u5199\u9762\u5F00"), "ok") : chip(t("\u5199\u9762\u5173\u65AD"), "esc")) + (m.canRevert ? chip(t("\u53EF\u64A4\u9500"), "acc") : "");
    const ovLine = m.override ? t('<div class="small faint" style="margin-top:6px">\u73B0\u884C\u8986\u76D6\uFF1Arevision ') + esc(fmtNum(m.override.revision)) + " \xB7 takeBudget " + esc(fmtNum(m.override.takeBudget)) + " \xB7 by " + esc(m.override.by || "\u2014") + (m.override.note ? t(" \xB7 \u5907\u6CE8 ") + esc(trunc(m.override.note, 60)) : "") + (m.override.updatedAt ? " \xB7 " + esc(relAge(m.override.updatedAt) || "") : "") + "</div>" : t('<div class="small faint" style="margin-top:6px">\u65E0\u8986\u76D6\uFF08\u751F\u6548\u503C=ContextPolicy \u7F3A\u7701\u2014\u2014PUT \u540E\u6B64\u5904\u73B0\u8BFB\u8986\u76D6\u884C\uFF09\u3002</div>');
    return t('<div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587\uFF08ContextPack \u5951\u7EA6\u9762\uFF09</span><span class="h-right">') + headChips + '</span></div><div class="card-b">' + sessionRowHtml(m) + '<div style="margin-top:10px">' + kv([
      ["takeBudget", '<b class="mono">' + esc(fmtNum(m.takeBudget)) + t('</b> <span class="small faint">\uFF08\u7F3A\u7701 ') + esc(fmtNum(m.takeBudgetDefault)) + t(" \xB7 \u57DF ") + m.policy.minTakeBudget + "\u2013" + m.policy.maxTakeBudget + "\uFF09</span>"],
      ["tagSig", m.tags.length ? m.tags.map((t2) => chip(t2, "mute")).join(" ") : t('<span class="faint small">\u672A\u88C5\u914D\uFF08tagSigSource=') + esc(m.tagSigSource) + "\uFF09</span>"],
      [t("\u5206\u7247\u7F13\u5B58"), m.stats ? t("\u5171 ") + esc(fmtNum(m.stats.total)) + t(" \u7247 \xB7 stale ") + esc(fmtNum(m.stats.stale)) + t(" \xB7 \u91CD\u5EFA\u9608\u503C ") + esc(fmtNum(m.stats.threshold)) : '<span class="faint">\u2014</span>'],
      ["shardMaxTokens", '<span class="mono">' + esc(fmtNum(m.policy.shardMaxTokens)) + "</span>"]
    ]) + "</div>" + ovLine + '<div data-ctx-notice></div><div style="margin:10px 0 4px">' + formHtml(m) + "</div>" + t('<h3 class="sec-title" style="margin:10px 0 6px">chunks\uFF08\u7D22\u5F15 TopK \u5019\u9009\uFF09</h3>') + chunksHtml(m) + t('<h3 class="sec-title" style="margin:10px 0 6px">\u76F8\u5173 Tag \u5206\u7247\uFF08TagCtxShard\uFF09</h3>') + shardsHtml(m) + (m.degraded && m.degradedReason ? '<div class="small faint" style="margin-top:8px">' + esc(m.degradedReason) + "</div>" : "") + "</div>";
  }
  function focusIn(box) {
    return !!(document.activeElement && box.contains(document.activeElement));
  }
  function paint(box, ctx2, id, model) {
    PANEL.set(id, { model });
    box.innerHTML = '<div class="card" data-ctx-card>' + renderSessionCtx(model) + "</div>";
    wireSessionCtx(box, ctx2, id);
  }
  async function refresh(ctx2, id, box) {
    const r = await ctx2.api.sessionContext(id);
    if (!r.ok || focusIn(box)) return;
    paint(box, ctx2, id, normalizeContext(r.data));
  }
  function wireSessionCtx(box, ctx2, id) {
    const q = (sel) => box.querySelector(sel);
    const notice2 = (html3) => {
      const slot = q("[data-ctx-notice]");
      if (slot) slot.innerHTML = html3;
    };
    const saveBtn = q("[data-ctx-save]");
    if (saveBtn) saveBtn.addEventListener("click", async () => {
      const cur = (PANEL.get(id) || {}).model;
      if (!cur) return;
      const built = buildSavePayload(cur, {
        takeBudget: q("[data-ctx-budget]").value,
        tagList: q("[data-ctx-tags]").value,
        note: q("[data-ctx-note]").value
      });
      if (!built.ok) {
        notice2('<div class="small faint">' + esc(built.error) + "</div>");
        return;
      }
      saveBtn.disabled = true;
      const r = await ctx2.api.saveSessionContext(id, built.payload);
      saveBtn.disabled = false;
      if (!r.ok) {
        notice2(t('<div class="small faint">\u4FDD\u5B58\u5931\u8D25\uFF08') + esc(String(r.error || r.status)) + "\uFF09" + (r.status === 409 ? t("\u2014\u2014\u4E50\u89C2\u9501\u51B2\u7A81\uFF1A\u8BF7\u91CD\u8FDB\u9762\u677F\u5BF9\u9F50 revision \u540E\u91CD\u8BD5\uFF08\u7A3F\u4E0D\u8986\u76D6\uFF0C\u4E0D\u9759\u9ED8\u91CD\u57FA\uFF09") : "") + "</div>");
        return;
      }
      notice2(t('<div class="small faint">\u5DF2\u4FDD\u5B58\uFF08revision ') + esc(String(r.data && r.data.override && r.data.override.revision || "?")) + "\uFF09" + (built.dropped.length ? t("\uFF1B\u8D85\u9650 tag \u6536\u655B ") + built.dropped.length + t(" \u4E2A\uFF1A") + esc(trunc(built.dropped.join(", "), 80)) : "") + "\u3002</div>");
      refresh(ctx2, id, box);
    });
    const revertBtn = q("[data-ctx-revert]");
    if (revertBtn) revertBtn.addEventListener("click", async () => {
      if (!window.confirm(t("\u64A4\u9500\u6700\u8FD1\u4E00\u6B21\u4E0A\u4E0B\u6587\u5206\u53C9\uFF08\u6062\u590D\u670D\u52A1\u7AEF\u5199\u524D\u5FEB\u7167\uFF09\uFF1F"))) return;
      revertBtn.disabled = true;
      const r = await ctx2.api.revertSessionContext(id);
      revertBtn.disabled = false;
      if (!r.ok) {
        notice2(t('<div class="small faint">\u64A4\u9500\u5931\u8D25\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u6808\u7A7A\u4E0D\u53EF\u64A4\uFF08\u4E0D\u5192\u5145\u64A4\u9500\u6210\u529F\uFF09\u3002</div>"));
        return;
      }
      notice2(t('<div class="small faint">\u5DF2\u64A4\u9500\u2014\u2014\u5199\u524D\u5FEB\u7167\u5DF2\u6062\u590D\u3002</div>'));
      refresh(ctx2, id, box);
    });
  }
  async function mountSessionCtx(el, ctx2, id) {
    const box = el.querySelector("[data-ctx-panel]");
    if (!box || !id) return;
    if (focusIn(box)) return;
    const cached = PANEL.get(id);
    if (cached && cached.model) box.innerHTML = '<div class="card" data-ctx-card>' + renderSessionCtx(cached.model) + "</div>";
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const r = await ctx2.api.sessionContext(id);
    if ((typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    if (!r.ok) {
      if (!focusIn(box)) {
        box.innerHTML = t('<div class="card"><div class="card-b"><div class="small faint">\u4E0A\u4E0B\u6587\u9762\u6682\u4E0D\u53EF\u8FBE\uFF08') + esc(String(r.error || r.status)) + (r.status === 503 ? t("\u2014\u2014contextidx \u672A\u88C5\u914D\uFF08\u9632\u5FA1\u9762\uFF09") : "") + t("\uFF09\u3002\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div></div></div>");
      }
      return;
    }
    if (focusIn(box)) return;
    paint(box, ctx2, id, normalizeContext(r.data));
  }

  // src/screens/chat.js
  function faces() {
    return typeof window !== "undefined" && window.__CC && window.__CC.chatFaces || null;
  }
  var THREAD = { sessionId: "", msgs: [] };
  var LAST = { route: "", paramKey: "", locale: "" };
  var SESS = { rows: null, ts: 0 };
  function sessAll(ctx2) {
    const snap = (ctx2.store.sessions() || []).map((s) => s && (s.id || s.sessionId) != null ? s : null).filter(Boolean);
    const live = Array.isArray(SESS.rows) ? SESS.rows : [];
    const seen = /* @__PURE__ */ new Set();
    return live.concat(snap).filter((s) => {
      const id = String(s && (s.id || s.sessionId) || "");
      if (!id || seen.has(id)) return false;
      seen.add(id);
      return true;
    });
  }
  async function hydrateSessLive(el, ctx2) {
    if (Date.now() - SESS.ts < 6e4 && SESS.rows) return;
    SESS.ts = Date.now();
    try {
      const r = await ctx2.api.sessions();
      if (!r.ok) return;
      SESS.rows = r.data && Array.isArray(r.data.sessions) ? r.data.sessions : [];
    } catch {
      return;
    }
    SESS.ts = Date.now();
    const cur = THREAD.sessionId;
    const sess = el.querySelector("[data-sess-list]");
    if (sess && el.isConnected) {
      sess.innerHTML = sessRowsHtml(ctx2, cur);
      bindSessRows(el, ctx2);
    }
  }
  function sessionOptions(ctx2, cur) {
    const rows = sessAll(ctx2).map((s) => String(s && (s.id || s.sessionId) || "")).filter(Boolean);
    const seen = /* @__PURE__ */ new Set(["root"]);
    const opts = rows.filter((id) => !seen.has(id)).slice(0, 30).map((id) => '<option value="' + esc(id) + '"' + (id === cur ? " selected" : "") + ">" + esc(trunc(id, 30)) + "</option>").join("");
    return '<option value="root"' + (cur === "root" ? " selected" : "") + ">" + esc(t("root \u4F1A\u8BDD\uFF08\u9ED8\u8BA4\uFF09")) + "</option>" + opts;
  }
  function composerHtml(ctx2, curSession) {
    const f = faces();
    const canSend = !!(f && typeof f.send === "function");
    return '<div class="cc-composer"><div class="cc-comp-row"><select class="inp cc-sess" data-cc-sess>' + sessionOptions(ctx2, curSession) + '</select></div><div class="cc-comp-row"><textarea class="inp cc-text" data-cc-text rows="2" placeholder="' + esc(t("\u63D0\u4EA4\u9700\u6C42\uFF0F\u627E\u6587\u6863\uFF0F\u66F4\u65B0\u6587\u6863\u2014\u2014\u4F8B\u5982\uFF1A\u8FDB\u884C\u4EE3\u7801\u5BA1\u6838\uFF0C\u770B\u770B\u5B9E\u9645\u548C\u6587\u6863\u5DEE\u8DDD\u6709\u591A\u5927\uFF0C\u5E76\u8FDB\u884C\u4FEE\u590D")) + '"></textarea><button class="btn primary cc-send" data-cc-send' + (canSend ? "" : " disabled") + ">" + esc(t("\u53D1\u9001")) + '</button></div><div class="small faint">' + (canSend ? esc(t("Enter \u53D1\u9001 \xB7 queue \u8BED\u4E49\uFF08\u9001\u8FBE agent \u4F1A\u8BDD\uFF1B\u7ED3\u8BBA\u7531 agent \u843D\u6587\u6863\u540E\u7ECF\u72B6\u6001\u5FBD\u4E0E\u8282\u70B9\u56FE\u53EF\u89C1\uFF09")) : esc(t("\u5BF9\u8BDD\u901A\u8DEF\u672A\u6CE8\u5165\uFF08chatfaces.js \u7F3A\u5E2D\uFF09\u2014\u2014\u9001\u8FBE\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u5192\u5145\u5DF2\u53D1\u9001"))) + "</div></div>";
  }
  function threadHtml(msgs, names) {
    if (!msgs.length) {
      return '<div class="card"><div class="ob-empty">' + esc(t("\u7EBF\u7A0B\u4E3A\u7A7A\u2014\u2014\u53D1\u7B2C\u4E00\u6761\u9700\u6C42\uFF1Bagent \u7684\u5B8C\u6574\u8F68\u8FF9\u5728\u539F\u751F\u754C\u9762\uFF08\u65B0\u7A97\u53E3\uFF09\u53EF\u89C1\u3002")) + "</div></div>";
    }
    return '<div class="card thread">' + msgs.map((m) => renderMessage(m, names)).join("") + "</div>";
  }
  function bindDocFull(el, ctx2) {
    el.querySelectorAll("[data-doc-full]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const ref = String(btn.getAttribute("data-doc-full") || "");
        const parts = ref.split("/");
        if (parts.length < 3) return;
        const [kind, scope, name] = parts;
        const r = await ctx2.api.object(decodeURIComponent(kind), decodeURIComponent(scope), decodeURIComponent(name));
        const body = r.ok && r.data && r.data.content != null ? String(r.data.content) : t("\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08") + (r.ok ? t("\u5185\u5BB9\u7F3A\u5E2D") : String(r.error || r.status)) + "\uFF09";
        let ov = document.querySelector("[data-doc-ov]");
        if (!ov) {
          ov = document.createElement("div");
          ov.setAttribute("data-doc-ov", "1");
          document.body.appendChild(ov);
        }
        ov.className = "doc-ov";
        ov.innerHTML = '<div class="doc-ov-card"><div class="doc-ov-h"><b class="mono">' + esc(decodeURIComponent(kind) + "/" + decodeURIComponent(scope) + "/" + decodeURIComponent(name)) + '</b><button class="btn sm" data-doc-ov-close>' + esc(t("\u5173\u95ED")) + '</button><button class="btn primary sm" data-doc-ov-edit>' + esc(t("\u5199\u5165\u8349\u7A3F\uFF08\u7F16\u8F91\u5668\uFF09")) + '</button></div><div class="doc-ov-b md-body">' + renderMarkdown(body) + "</div></div>";
        ov.querySelector("[data-doc-ov-close]").addEventListener("click", () => {
          ov.className = "doc-ov hidden";
        });
        ov.querySelector("[data-doc-ov-edit]").addEventListener("click", () => {
          ctx2.go("doc/" + encodeURIComponent(decodeURIComponent(kind)) + "/" + encodeURIComponent(decodeURIComponent(scope)) + "/" + encodeURIComponent(decodeURIComponent(name)));
          ov.className = "doc-ov hidden";
        });
      });
    });
  }
  function bindComposer(el, ctx2, rootEl) {
    const sendBtn = rootEl.querySelector("[data-cc-send]");
    const sessSel = rootEl.querySelector("[data-cc-sess]");
    const textTa = rootEl.querySelector("[data-cc-text]");
    if (!sendBtn || !textTa) return;
    const doSend = async () => {
      const text2 = textTa.value.trim();
      if (!text2) return;
      const sessionId = sessSel && sessSel.value || "root";
      THREAD.sessionId = sessionId;
      const f = faces();
      if (!f || typeof f.send !== "function") return;
      THREAD.msgs.push({ role: "user", text: text2, ts: (/* @__PURE__ */ new Date()).toISOString().slice(11, 19) });
      textTa.value = "";
      paintThread(rootEl, ctx2);
      const r = await f.send(sessionId, text2);
      if (r && r.status === 200) {
        THREAD.msgs.push({
          role: "assistant",
          name: "rootAgent",
          seq: r.data && r.data.requestId || "",
          ts: (/* @__PURE__ */ new Date()).toISOString().slice(11, 19),
          text: t("\u5DF2\u9001\u8FBE\uFF08queue\uFF09\u2014\u2014agent \u4F1A\u8BDD\u5DF2\u53D7\u7406\u3002\u7ED3\u8BBA\u5C06\u7531 agent \u751F\u6210\u5E76\u81EA\u884C\u843D\u6587\u6863\uFF1B\u6587\u6863\u72B6\u6001\u5FBD\u4E0E workflow \u8282\u70B9\u56FE\u968F\u540E\u53EF\u89C1\u3002\u539F\u751F\u8F68\u8FF9\uFF1A"),
          tool: "send \u2192 sessionController.prompt \xB7 requestId=" + (r.data && r.data.requestId || "\u2014")
        });
      } else {
        THREAD.msgs.push({
          role: "assistant",
          name: "system",
          text: t("\u9001\u8FBE\u5931\u8D25\uFF08") + (r && r.status || "\u2014") + "\uFF09" + (r && r.data && r.data.error ? " \xB7 " + r.data.error : "")
        });
      }
      paintThread(rootEl, ctx2);
    };
    sendBtn.addEventListener("click", doSend);
    textTa.addEventListener("keydown", (ev) => {
      if (ev.key === "Enter" && !ev.shiftKey) {
        ev.preventDefault();
        doSend();
      }
    });
  }
  function paintThread(rootEl, ctx2) {
    const box = rootEl.querySelector("[data-thread]");
    if (!box) return;
    const names = new Set((ctx2.store.objects() || []).map((o) => o.name).filter(Boolean));
    box.innerHTML = threadHtml(THREAD.msgs, names);
    box.scrollTop = box.scrollHeight;
    bindDocFull(box, ctx2);
  }
  function nativeSrc(target) {
    let origin = "/";
    try {
      origin = location.origin + "/";
    } catch {
    }
    return origin + "?embedded=1#/chat/" + (target ? encodeURIComponent(String(target)) : "");
  }
  function sessRowsHtml(ctx2, cur) {
    const rows = sessAll(ctx2).map((s) => {
      const id = String(s && (s.id || s.sessionId) || "");
      if (!id) return "";
      const state = String(s && s.state || "");
      const beat = String(s && s.lastBeat || "");
      const sel = id === cur ? ' style="outline:1px solid var(--accent);background:var(--panel-2);border-radius:8px"' : "";
      return '<div class="dash-row" data-sess-go="' + esc(id) + '" role="link"' + sel + ' title="' + esc(trunc(String(s && s.lastAction || ""), 120)) + '"><div class="dr-main"><b class="mono">' + esc(trunc(id, 24)) + '</b></div><div class="dr-side">' + (state ? chip(state, state === "escalated" || state === "suspected-stall" ? "esc" : "acc") : "") + (beat ? '<span class="small faint">' + esc(relAge(beat) || "") + "</span>" : "") + "</div></div>";
    }).filter(Boolean).join("");
    return rows || '<div class="small faint" style="padding:4px 6px">' + esc(t("\u5FEB\u7167\u65E0\u6D3B\u6027\u4F1A\u8BDD\u884C\u2014\u2014\u53EF\u5728\u53F3\u4FA7\u539F\u751F\u9762\u76F4\u63A5\u5BF9\u8BDD\uFF08\u65B0\u4F1A\u8BDD\u7531\u5BBF\u4E3B\u539F\u751F\u5EFA\u6863\uFF09")) + "</div>";
  }
  function agentRowsHtml(ctx2) {
    const rows = (ctx2.store.objects() || []).filter((o) => o && o.kind === "Agent" && o.name).map((o) => '<div class="dash-row" data-go="' + esc("chat/" + encodeURIComponent(String(o.name))) + '" role="link"><div class="dr-main"><b class="mono">' + esc(trunc(String(o.name), 24)) + '</b></div><div class="dr-side"><span class="small faint">' + esc(String(o.scope || "global")) + "</span></div></div>").join("");
    return rows || '<div class="small faint" style="padding:4px 6px">' + esc(t("\u5FEB\u7167\u6682\u65E0 Agent \u5BF9\u8C61\u679A\u4E3E")) + "</div>";
  }
  function ctxBarHtml(sessionId) {
    if (!sessionId) return "";
    return '<details class="ex-block" style="margin:0 0 10px;background:var(--panel-2)"><summary class="ex-sum"><span class="ex-chev">\u25B6</span>' + esc(t("\u4E0A\u4E0B\u6587\uFF08ContextPack \u5951\u7EA6\u9762\uFF09")) + '</summary><div class="ex-body"><div data-ctx-panel><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u4E0A\u4E0B\u6587\u6CE8\u6C34\u4E2D\u2026\uFF08GET /v1/sessions/")) + esc(encodeURIComponent(sessionId)) + "/context\uFF09</div></div></div></div></div></details>";
  }
  function faceHtml(target) {
    return '<div style="position:relative"><div data-native-ph style="position:absolute;inset:0;display:flex;align-items:center;justify-content:center;border:1px dashed var(--border);border-radius:12px;background:var(--panel-2);z-index:2;pointer-events:none"><div class="small faint">' + esc(t("\u539F\u751F\u4F1A\u8BDD\u9762\u88C5\u8F7D\u4E2D\u2026")) + '</div></div><iframe data-native-iframe data-target="' + esc(target) + '" title="' + esc(t("\u5BBF\u4E3B\u539F\u751F\u4F1A\u8BDD\u9762\uFF08\u771F\u5B9E\u8F6C\u5F55\u4E0E\u53D1\u9001\uFF09")) + '" src="' + esc(nativeSrc(target)) + '" referrerpolicy="same-origin" style="width:100%;height:calc(100vh - 96px);border:1px solid var(--border);border-radius:12px;background:var(--panel-2);display:block"></iframe></div>';
  }
  function bindFace(el) {
    const ifr = el.querySelector("[data-native-iframe]");
    const ph = el.querySelector("[data-native-ph]");
    if (ifr && ph) ifr.addEventListener("load", () => {
      ph.style.display = "none";
    });
  }
  function swapFace(el, target) {
    const ifr = el.querySelector("[data-native-iframe]");
    if (!ifr) return;
    const src = nativeSrc(target);
    if (ifr.getAttribute("src") !== src) {
      ifr.setAttribute("src", src);
      const ph = el.querySelector("[data-native-ph]");
      if (ph) ph.style.display = "";
    }
    ifr.setAttribute("data-target", String(target));
  }
  function bindSessRows(el, ctx2) {
    el.querySelectorAll("[data-sess-go]").forEach((row) => {
      row.addEventListener("click", () => {
        const id = String(row.getAttribute("data-sess-go") || "");
        if (!id) return;
        THREAD.sessionId = id;
        swapFace(el, id);
        const listBox = el.querySelector("[data-sess-list]");
        if (listBox) {
          listBox.innerHTML = sessRowsHtml(ctx2, id);
          bindSessRows(el, ctx2);
        }
        ensureCtxBar(el, ctx2, id);
      });
    });
  }
  function ensureCtxBar(el, ctx2, id) {
    const holder = el.querySelector("[data-ctx-holder]");
    if (!holder) return;
    if (!holder.querySelector("[data-ctx-panel]")) holder.innerHTML = ctxBarHtml(id);
    mountSessionCtx(el, ctx2, id);
  }
  function bindPop(el, target) {
    const pop = el.querySelector("[data-native-pop]");
    if (pop) pop.addEventListener("click", () => {
      try {
        window.open(nativeSrc(target), "_blank");
      } catch {
      }
    });
  }
  function dialogScreen(el, ctx2, target) {
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u4F1A\u8BDD"),
      t("\u539F\u751F DeepSeek \u4F1A\u8BDD\u9762\u2014\u2014\u771F\u5B9E\u8F6C\u5F55\u4E0E\u53D1\u9001\uFF1B\u5DE6\u4FA7\u5207\u6362\u4F1A\u8BDD"),
      chip(t("\u65B0\u7A97\u53E3\u539F\u751F\u754C\u9762"), "mute", false, 'data-native-pop="1" style="cursor:pointer"')
    ) + '<div style="display:grid;grid-template-columns:300px minmax(0,1fr);gap:14px;align-items:start"><aside class="card" style="padding:8px;max-height:calc(100vh - 96px);overflow:auto"><div class="small faint" style="padding:2px 6px 8px;font-weight:700">' + esc(t("\u4F1A\u8BDD\uFF08\u6D3B\u6027\u8D26\uFF09")) + "</div><div data-sess-list>" + sessRowsHtml(ctx2, target) + '</div><div class="small faint" style="padding:12px 6px 8px;font-weight:700">' + esc(t("Agent")) + "</div><div data-agent-list>" + agentRowsHtml(ctx2) + '</div></aside><section style="min-width:0"><div data-ctx-holder>' + ctxBarHtml(THREAD.sessionId) + "</div>" + faceHtml(target) + "</section></div></div>";
    bindFace(el);
    bindSessRows(el, ctx2);
    bindPop(el, target);
    if (THREAD.sessionId) mountSessionCtx(el, ctx2, THREAD.sessionId);
    hydrateSessLive(el, ctx2);
  }
  function render(el, ctx2) {
    const id = ctx2.params && ctx2.params.id;
    const agent = ctx2.params && ctx2.params.agent;
    const route = String(ctx2.route || "");
    const paramKey = id != null && id !== "" ? "session/" + String(id) : "chat/" + String(agent || "");
    const locale2 = getLocale();
    const fresh = LAST.route !== route || LAST.paramKey !== paramKey || LAST.locale !== locale2;
    LAST.route = route;
    LAST.paramKey = paramKey;
    LAST.locale = locale2;
    if (fresh) {
      if (id != null && id !== "") {
        THREAD.sessionId = String(id);
      } else if (agent != null && agent !== "") {
        THREAD.sessionId = sessKnown(ctx2, String(agent)) ? String(agent) : "";
      }
    } else if (!THREAD.sessionId && agent != null && agent !== "" && sessKnown(ctx2, String(agent))) {
      THREAD.sessionId = String(agent);
      dialogScreen(el, ctx2, THREAD.sessionId);
      return;
    }
    const target = THREAD.sessionId || String(agent || id || "");
    const face = el.querySelector("[data-native-iframe]");
    if (!fresh && face && face.getAttribute("data-target") === target) {
      const sess = el.querySelector("[data-sess-list]");
      if (sess) {
        sess.innerHTML = sessRowsHtml(ctx2, target);
        bindSessRows(el, ctx2);
      }
      const ag = el.querySelector("[data-agent-list]");
      if (ag) ag.innerHTML = agentRowsHtml(ctx2);
      return;
    }
    dialogScreen(el, ctx2, target);
  }
  function sessKnown(ctx2, candidate) {
    return sessAll(ctx2).some((s) => String(s && (s.id || s.sessionId) || "") === candidate);
  }

  // src/screens/welcome.js
  function statCard(label, num, sub, route, cls) {
    return '<div class="stat ' + esc(cls || "") + '" data-go="' + esc(route) + '" role="link"><div class="st-num">' + esc(fmtNum(num)) + '</div><div class="st-label">' + esc(label) + '<span class="small faint">' + esc(sub || "") + "</span></div></div>";
  }
  function loopCardHtml(data) {
    const d = data && typeof data === "object" ? data : {};
    const fields = [
      [t("\u5FAA\u73AF"), d.loop != null ? d.loop : d.id],
      [t("\u76F8\u4F4D"), d.phase || d.state],
      [t("\u8D1F\u8D23"), d.agent || d.route],
      [t("\u5F00\u59CB\u4E8E"), d.startedAt ? relAge(d.startedAt) || String(d.startedAt) : d.started_at ? relAge(d.started_at) || String(d.started_at) : null],
      [t("\u89E6\u53D1"), d.trigger || d.reason]
    ].filter(([, v]) => v != null && v !== "");
    const chips = fields.map(([k, v]) => chip(k + ": " + v, "mute")).join("");
    return '<div class="card report-loop"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u5FAA\u73AF"), "acc") + chips + '<span class="report-loop-acts"><span data-ignite-note></span><button class="btn sm" data-ignite="1">' + esc(t("\u89E6\u53D1\u4E8B\u4EF6")) + "</button></span></div></div></div>";
  }
  function bindIgnite(el, ctx2) {
    const btn = el.querySelector("[data-ignite]");
    if (!btn) return;
    btn.addEventListener("click", async () => {
      if (btn.disabled) return;
      const note = el.querySelector("[data-ignite-note]");
      btn.disabled = true;
      btn.textContent = t("\u70B9\u706B\u4E2D\u2026");
      const r = await ctx2.api.eventPub({
        type: "LoopRequested",
        involved: "ServiceDesc/core/core_system",
        dedupKey: "console-ignite-" + Date.now()
      });
      if (!el.isConnected) return;
      if (r && r.ok) {
        if (note) note.innerHTML = chip(t("\u5DF2\u6CE8\u5165 LoopRequested\uFF08queue \u8BED\u4E49\u2014\u2014\u6865\u63A5\u9762\u62C9\u8D77\u5FAA\u73AF\uFF09"), "ok");
        setTimeout(() => {
          if (el.isConnected) hydrateLoop(el, ctx2);
        }, 3e3);
      } else {
        btn.disabled = false;
        btn.textContent = t("\u89E6\u53D1\u4E8B\u4EF6");
        if (note) note.innerHTML = chip(
          t("\u70B9\u706B\u5931\u8D25\uFF08") + String(r && r.status || "\u2014") + (r && r.error ? " " + String(r.error) : "") + t("\uFF09\u2014\u2014\u53EF\u91CD\u8BD5"),
          "danger"
        );
      }
    });
  }
  async function hydrateLoop(el, ctx2) {
    const box = el.querySelector("[data-loop-box]");
    if (!box) return;
    const r = await ctx2.api.loopCurrent();
    if (!el.isConnected || box.getAttribute("data-loop-box") === "stale") return;
    if (!r.ok) {
      box.innerHTML = '<div class="card report-loop" data-report-loop="degraded"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u5FAA\u73AF"), "mute") + t('<span class="small faint">\u5FAA\u73AF\u72B6\u6001\u6682\u4E0D\u53EF\u8FBE\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u4E0D\u5192\u5145\u65E0\u5FAA\u73AF\u3002</span></div></div></div>");
      return;
    }
    if (!r.data || !r.data.loop && !r.data.id && !r.data.phase) {
      box.innerHTML = '<div class="card report-loop" data-report-loop="idle"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF"), "mute") + t('<span class="small faint">\u672C\u5468\u671F\u6682\u65E0\u5FAA\u73AF\u8BB0\u5F55\u3002\u70B9\u706B\u4E0B\u4E00\u8F6E\uFF1A<code class="mono">corectl loop</code></span></div></div></div>');
      return;
    }
    box.innerHTML = loopCardHtml(r.data);
    bindIgnite(el, ctx2);
  }
  function zoneHead(label, route) {
    return '<div class="zone-h"><span class="zone-t">' + esc(label) + "</span>" + (route ? chipGo(t("\u5168\u90E8"), route, "mute", false) : "") + "</div>";
  }
  function runningWorkflowsHtml(ctx2) {
    const runs = ctx2.store.runs().filter((r) => {
      const st = String(r && (r.status || r.state) || "");
      return st && st !== "done" && st !== "failed" && st !== "canceled";
    }).slice(0, 8);
    const rows = runs.map((r) => {
      const wf = String(r.workflow || r.wf || "\u2014");
      const phase = String(r.phase || r.state || "\u2014");
      const st = String(r.status || "");
      const started = r.startedAt ? relAge(r.startedAt) || "" : "";
      const route = r.workflow ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(r.run != null ? r.run : "")) : "workflows";
      return '<div class="dash-row" data-go="' + esc(route) + '" role="link"><div class="dr-main"><b class="mono">' + esc(trunc(wf, 26)) + '</b><span class="small faint">#' + esc(String(r.run != null ? r.run : "\u2014")) + '</span></div><div class="dr-side">' + chip(phase, st === "running" ? "acc" : "mute") + (started ? '<span class="small faint">' + esc(started) + "</span>" : "") + "</div></div>";
    }).join("");
    return '<div class="card">' + (rows || '<div class="ob-empty">' + esc(t("\u5F53\u524D\u65E0\u8FDB\u884C\u4E2D\u5FAA\u73AF")) + "</div>") + "</div>";
  }
  function runOwnerScope(r) {
    const inv = String(r && (r.involved || r.route || "") || "");
    const m = /^Workflow\/([^/\s]+)\//.exec(inv);
    if (m) return m[1];
    return String(r && (r.scope || r.ownerScope) || "") || "global";
  }
  function pendingRunsOf(ctx2) {
    return ctx2.store.runs().filter((r) => r && (String(r.status || r.state || "") === "suspended" || String(r.phase || "") === "AwaitingApproval"));
  }
  function pendingRunRowHtml(x) {
    const btnAttrs = (t2) => " data-appr-" + t2 + '="1" data-appr-wf="' + esc(x.wf) + '" data-appr-scope="' + esc(x.scope) + '" data-appr-run="' + esc(x.run) + '" data-appr-key="' + esc(x.key) + '"';
    return '<div class="dash-row" data-appr-row><div class="dr-main">' + chip(t("\u5F85\u6279"), "warn") + '<b class="mono">' + esc(trunc(x.wf, 22)) + '</b><span class="small faint">#' + esc(x.run) + "</span>" + chip(x.phase, "warn") + (x.started ? '<span class="small faint">' + esc(x.started) + "</span>" : "") + '</div><div class="dr-side" style="display:flex;gap:6px;align-items:center"><button class="btn sm primary"' + btnAttrs("grant") + ">" + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button><button class="btn sm"' + btnAttrs("reject") + ">" + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button></div></div>";
  }
  function bindApprovalGates(el, ctx2) {
    const act = async (btn, type2) => {
      if (btn.disabled) return;
      const wf = btn.getAttribute("data-appr-wf") || "";
      const scope = btn.getAttribute("data-appr-scope") || "global";
      const run = btn.getAttribute("data-appr-run") || "";
      const key = btn.getAttribute("data-appr-key") || run;
      const ask = type2 === "ApprovalGranted" ? t("\u786E\u8BA4\u6279\u51C6\u7EE7\u7EED\uFF1Frun \u5C06\u6CBF Workflow \u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\u7EED\u884C\uFF1A") : t("\u786E\u8BA4\u9A73\u56DE\u7EC8\u6B62\uFF1Frun \u5C06\u8FDB\u5165\u7EC8\u6001 Failed\uFF08\u4E0D\u53EF\u6062\u590D\uFF09\uFF1A");
      if (!window.confirm(ask + wf + " #" + run)) return;
      btn.disabled = true;
      const r = await ctx2.api.eventPub({
        type: type2,
        involved: "Workflow/" + (scope || "global") + "/" + wf,
        dedupKey: "console-" + key
      });
      if (r && r.ok) {
        const row = btn.closest("[data-appr-row]");
        if (row) row.style.display = "none";
        ctx2.api.snapshot().then((s) => {
          if (s && s.ok) ctx2.store.init(s.data);
        });
      } else {
        btn.disabled = false;
        const verb = type2 === "ApprovalGranted" ? t("\u6279\u51C6\u5931\u8D25") : t("\u9A73\u56DE\u5931\u8D25");
        window.alert(verb + t("\uFF08") + String(r && r.status || "\u2014") + (r && r.error ? " " + String(r.error) : "") + t("\uFF09\u2014\u2014run \u4ECD\u5728\u7B49\u5BA1\u6279\uFF0C\u53EF\u91CD\u8BD5\u3002"));
      }
    };
    el.querySelectorAll("[data-appr-grant]").forEach((btn) => btn.addEventListener("click", (ev) => {
      ev.stopPropagation();
      act(btn, "ApprovalGranted");
    }));
    el.querySelectorAll("[data-appr-reject]").forEach((btn) => btn.addEventListener("click", (ev) => {
      ev.stopPropagation();
      act(btn, "ApprovalRejected");
    }));
  }
  function pendingApprovalsHtml(ctx2) {
    const runRows = pendingRunsOf(ctx2).slice(0, 8).map((r) => {
      const wf = String(r.workflow || r.wf || "");
      if (!wf) return "";
      const run = String(r.run != null ? r.run : "");
      const key = r.eventId != null && r.eventId !== "" ? String(r.eventId) : run;
      return pendingRunRowHtml({
        wf,
        run,
        key,
        scope: runOwnerScope(r),
        phase: String(r.phase || r.state || "AwaitingApproval"),
        started: r.startedAt ? relAge(r.startedAt) || "" : ""
      });
    }).filter(Boolean);
    const drafts = ctx2.store.drafts().filter((d) => d && d.kind && d.name).map((d) => ({
      src: "draft",
      t: d.updatedAt || d.ts || null,
      go: "doc/" + encodeURIComponent(d.kind) + "/" + encodeURIComponent(d.scope || "global") + "/" + encodeURIComponent(d.name),
      title: d.kind + "/" + (d.scope || "global") + "/" + d.name,
      tag: t("\u7A3F")
    }));
    const evs = ctx2.store.events().filter((e) => e && (e.type === "ApprovalNeeded" || e.type === "Escalate")).slice(0, 8).map((e) => ({
      src: "ev",
      t: e.ts || null,
      go: "event/" + encodeURIComponent(String(e.id ?? "")),
      title: e.type + " \xB7 " + String(e.involved || ""),
      tag: e.type === "ApprovalNeeded" ? t("\u5F85\u5BA1") : t("\u5347\u7EA7")
    }));
    const rest = drafts.concat(evs).sort((a, b) => String(b.t || "").localeCompare(String(a.t || ""))).slice(0, 8).map((x) => '<div class="dash-row" data-go="' + esc(x.go) + '" role="link"><div class="dr-main">' + chip(x.tag, x.src === "draft" ? "warn" : "esc") + '<b class="mono">' + esc(trunc(x.title, 34)) + "</b></div>" + (x.t ? '<div class="dr-side"><span class="small faint">' + esc(relAge(x.t) || "") + "</span></div>" : "") + "</div>").join("");
    const rows = runRows.join("") + rest;
    if (!rows) return '<div class="card"><div class="ob-empty">' + esc(t("\u5F85\u5BA1\u533A\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09")) + "</div></div>";
    return '<div class="card">' + rows + "</div>";
  }
  var AG_SESS = { list: null, at: 0 };
  var AG_SESS_TTL_MS = 6e4;
  async function sessionsCached(ctx2) {
    if (AG_SESS.list && Date.now() - AG_SESS.at < AG_SESS_TTL_MS) return AG_SESS.list;
    const r = await ctx2.api.sessions();
    if (r && r.ok && Array.isArray(r.data && r.data.sessions)) {
      AG_SESS = { list: r.data.sessions, at: Date.now() };
      return AG_SESS.list;
    }
    return null;
  }
  function sessStateCls(st) {
    return st === "escalated" ? "esc" : st === "running" ? "acc" : st === "suspected-stall" ? "warn" : "mute";
  }
  function sessHitOf(list2, name) {
    const byId = list2.find((s) => String(s && s.sessionId || "").includes(name));
    if (byId) return byId;
    return list2.find((s) => {
      try {
        return JSON.stringify([s && s.involved, s && s.route, s && s.agent]).includes(name);
      } catch {
        return false;
      }
    });
  }
  function sessLiveHtml(hit) {
    if (!hit) return chip(t("\u65E0\u4F1A\u8BDD"), "mute");
    const st = String(hit.state || "");
    const sub = [
      String(hit.phase || ""),
      trunc(String(hit.lastAction || ""), 18),
      hit.lastBeat ? relAge(hit.lastBeat) || "" : ""
    ].filter(Boolean).join(" \xB7 ");
    return (st ? chip(st, sessStateCls(st)) : "") + (sub ? '<span class="small faint">' + esc(sub) + "</span>" : "");
  }
  async function hydrateAgentLive(el, ctx2) {
    const slots = el.querySelectorAll("[data-agent-live]");
    if (!slots.length) return;
    const list2 = await sessionsCached(ctx2);
    if (!el.isConnected) return;
    if (!list2) {
      slots.forEach((slot) => {
        slot.innerHTML = chip(t("\u4F1A\u8BDD\u6D3B\u6027\u6682\u4E0D\u53EF\u8BFB"), "mute");
      });
      return;
    }
    slots.forEach((slot) => {
      const name = String(slot.getAttribute("data-agent-live") || "");
      if (!name) return;
      slot.innerHTML = sessLiveHtml(sessHitOf(list2, name));
    });
  }
  function agentsHtml(ctx2, filterAgents) {
    const objs = ctx2.store.objects();
    let agents = objs.filter((o) => o.kind === "Agent").map((o) => ({ name: o.name, scope: o.scope || "global", route: objRoute("Agent", o.scope || "global", o.name) }));
    if (filterAgents && filterAgents.length) agents = agents.filter((a) => filterAgents.includes(a.name));
    const sessions = ctx2.store.sessions();
    const rows = agents.map((a) => {
      const sess = sessions.find((s) => String(s && (s.agent || s.route) || "").includes(a.name)) || sessions.find((s) => String(s && s.id || "") === a.name);
      const state = sess ? String(sess.state || sess.status || "") : "";
      return '<div class="dash-row" data-go="' + esc("chat/" + encodeURIComponent(a.name)) + '" role="link"><div class="dr-main"><b class="mono">' + esc(trunc(a.name, 30)) + '</b><span class="small faint">' + esc(a.scope) + '</span></div><div class="dr-side" data-agent-live="' + esc(a.name) + '">' + (state ? chip(state, state === "escalated" || state === "suspected-stall" ? "esc" : "acc") : chip(t("\u65E0\u4F1A\u8BDD"), "mute")) + "</div></div>";
    }).join("");
    if (!rows) return '<div class="card"><div class="ob-empty">' + esc(t("\u5FEB\u7167\u6682\u65E0\u5BF9\u8C61\u679A\u4E3E")) + "</div></div>";
    return '<div class="card">' + rows + "</div>";
  }
  function dashLayout(ctx2, opts) {
    const store2 = ctx2.store;
    const inbox = store2.inbox();
    const inboxCard = inbox ? statCard(
      t("\u6536\u4EF6\u7BB1"),
      Array.isArray(inbox) ? inbox.length : inbox && inbox.items && inbox.items.length || 0,
      t("\u5F85\u4F60\u5904\u7406"),
      "inbox",
      ""
    ) : "";
    const stats = statCard(t("\u5728\u518C\u5BF9\u8C61"), (store2.objects() || []).length, t("\u5168\u90E8 kind \u5408\u8BA1"), "obj", "") + statCard(t("\u4E8B\u4EF6"), (store2.events() || []).length, t("\u5F52\u6863\u884C"), "events", "") + statCard(t("\u5F85\u5BA1"), (store2.drafts() || []).length + (store2.events() || []).filter((e) => e && e.type === "ApprovalNeeded").length, t("\u7A3F\u222A\u5F85\u5BA1\u4E8B\u4EF6"), "approvals", "warn") + inboxCard + statCard(t("\u8FDB\u884C\u4E2D\u4EFB\u52A1"), (store2.tasks() || []).filter((x) => ["running", "verifying"].includes(String(x && x.status && (x.status.phase || x.status) || ""))).length, t("\u4EFB\u52A1\u4E2D\u5FC3"), "tasks", "");
    return '<div class="view-inner">' + (opts.head || "") + (opts.composer ? '<div class="dash-composer">' + composerHtml(ctx2, "") + '<div class="small faint"><span class="ref-chip" data-go="chat/core-headless">' + esc(t("\u6253\u5F00\u5BF9\u8BDD\u7EBF\u7A0B")) + "</span></div></div>" : "") + '<div class="dash-stats"' + (inboxCard ? "" : ' style="grid-template-columns:repeat(4,1fr)"') + ">" + stats + '</div><div data-loop-box><div class="card report-loop"><div class="card-b"><div class="report-loop-top">' + chip(t("\u5F53\u524D\u5FAA\u73AF"), "mute") + t('<span class="small faint">\u5FAA\u73AF\u72B6\u6001\u73B0\u8BFB\u4E2D\u2026</span>') + '</div></div></div></div><div class="dash-grid"><div class="dash-zone">' + zoneHead(t("\u6B63\u5728\u8DD1\u7684 Workflows"), "workflows") + runningWorkflowsHtml(ctx2) + '</div><div class="dash-zone">' + zoneHead(t("\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279"), "approvals") + pendingApprovalsHtml(ctx2) + '</div><div class="dash-zone">' + zoneHead(t("\u6838\u5FC3 Agent"), "tree") + agentsHtml(ctx2, opts.coreAgents) + "</div></div>" + (opts.foot || "") + '<div class="w-foot small faint">' + t('<div class="w-foot small faint">\u6570\u636E\u9762\uFF1AGET /v1/snapshot \u57FA\u7EBF + /v1/stream patch \u589E\u91CF\uFF08WS \u65AD\u7EBF\u81EA\u52A8\u964D\u7EA7\u8F6E\u8BE2\uFF09\uFF1B') + t("\uFF08rv \u672A\u53D8\u4E0D\u91CD\u8BFB\u6B63\u6587\u2014\u2014\u964D\u7EA7\u8F6E\u8BE2\u8DEF\u5F84\u540C\u6837\u751F\u6548\uFF09</span>") + "</div>";
  }
  function render2(el, ctx2) {
    el.innerHTML = dashLayout(ctx2, {
      head: '<div class="dash-head"><h2 class="page-title">' + esc(t("\u603B\u89C8")) + '</h2><span class="small faint">' + esc(t("\u7CFB\u7EDF\u89C6\u89D2\u5927\u5C4F\u2014\u2014\u6B63\u5728\u8DD1\u7684\u4EFB\u52A1\u3001\u6700\u8FD1\u9700\u8981\u7684\u5BA1\u6279\u3001\u6838\u5FC3 Agent \u4E00\u773C\u53EF\u8BFB")) + "</span></div>",
      composer: true
    });
    bindComposer(el, ctx2, el);
    hydrateLoop(el, ctx2);
    bindApprovalGates(el, ctx2);
    hydrateAgentLive(el, ctx2);
  }
  function renderService(el, ctx2) {
    const name = String(ctx2.params.name || "");
    const sd = ctx2.store.objects().find((o) => o.kind === "ServiceDesc" && o.name === name);
    let agents = [];
    if (sd && typeof sd.content === "string") {
      agents = [...String(sd.content).matchAll(/agent:\/\/([A-Za-z0-9_.-]+)/g)].map((m) => m[1]);
    }
    el.innerHTML = dashLayout(ctx2, {
      head: '<div class="dash-head"><h2 class="page-title">' + esc(t("Service")) + " \xB7 " + esc(name) + '</h2><span class="small faint">' + esc(t("Service \u89C6\u89D2\u5927\u5C4F\uFF08\u4E0E\u7CFB\u7EDF\u5927\u5C4F\u540C\u6784\u5E03\u5C40\u2014\u2014\u540E\u7EED ServiceAgent \u767B\u5F55\u5373\u843D\u6B64\u89C6\u56FE\uFF09")) + "</span></div>",
      coreAgents: agents.length ? agents : null,
      foot: '<div class="card"><div class="card-b"><div class="small faint">' + (sd ? esc(t("ServiceDesc \u5728\u518C\u2014\u2014agents=")) + esc(agents.join(", ") || "\u2014") : esc(t("ServiceDesc \u5BF9\u8C61\u4E0D\u5728\u518C\uFF08\u8BCD\u5F62\u6295\u5F71\u81EA\u5FEB\u7167\u2014\u2014\u4E0D\u5192\u5145\u5728\u518C\uFF09"))) + "</div></div></div>"
    });
    hydrateLoop(el, ctx2);
    bindApprovalGates(el, ctx2);
    hydrateAgentLive(el, ctx2);
  }

  // src/screens/workflows.js
  function wfAnchorOfRoute(route) {
    const r = String(route || "");
    if (!r) return "";
    if (r.startsWith("workflows/")) return r.slice("workflows/".length);
    if (/-workflow$/.test(r)) return "core-workflow";
    return "";
  }
  function phaseGraphHtml(run, ctx2) {
    const phases = Array.isArray(run.phases) ? run.phases : [];
    const cur = String(run.phase || "");
    const st = String(run.status || "");
    const nodeHtml = (p, i) => {
      const active = p === cur;
      const done = st === "done" || cur && phases.indexOf(cur) > i;
      const cls = "pg-node" + (active ? " active" : "") + (done ? " done" : "") + (p === "Failed" ? " fail" : "");
      return '<div class="' + cls + '" data-pg-node="' + esc(p) + '"><div class="pg-dot"></div><div class="pg-name">' + esc(p) + "</div>" + (active ? chip(st || t("\u8FDB\u884C\u4E2D"), "acc") : "") + "</div>";
    };
    const body = phases.length ? phases.map((p, i) => nodeHtml(p, i)).join('<span class="pg-edge">\u2192</span>') : '<div class="pg-node active"><div class="pg-dot"></div><div class="pg-name">' + esc(cur || t("\u5355\u8282\u70B9\uFF08\u58F0\u660E\u672A\u6295\u5F71\u76F8\u4F4D\uFF09")) + "</div>" + runChip(run) + "</div>";
    return '<div class="pg-wrap">' + body + '</div><div class="small faint">' + esc(t("\u70B9\u51FB\u8282\u70B9\u770B\u8BE5\u76F8\u4F4D\u7684\u4E8B\u4EF6\u94FE\u4E0E\u6D89\u53CA\u6587\u6863\uFF1B\u63A8\u8FDB\u7531 patch \u589E\u91CF\u5B9E\u65F6\u5237\u65B0\u3002")) + '</div><div data-pg-drawer style="margin-top:10px"></div>';
  }
  function bindPhaseGraph(el, ctx2, run) {
    el.querySelectorAll("[data-pg-node]").forEach((n) => {
      n.addEventListener("click", () => {
        const ph = String(n.getAttribute("data-pg-node") || "");
        const drawer = el.querySelector("[data-pg-drawer]");
        if (!drawer) return;
        const evs = ctx2.store.events().filter((e) => {
          const hay = JSON.stringify([e.dedupKey, e.involved, e.type]);
          return hay.includes(String(run.workflow || "")) || hay.includes(ph);
        }).slice(0, 10);
        const involved = String(run.involved || "");
        const evRows = evs.length ? evs.map((e) => '<div class="mini-row"><div class="mr-main"><div><b>' + esc(String(e.type || "")) + '</b></div><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 60)) + '</div></div><span class="small faint">' + esc(relAge(e.ts) || "") + "</span></div>").join("") : '<div class="small faint">' + esc(t("\u7A97\u53E3\u5185\u65E0\u8BE5\u76F8\u4F4D\u4E8B\u4EF6\uFF08\u8BDA\u5B9E\u7A7A\u6001\uFF09\u3002")) + "</div>";
        drawer.innerHTML = '<div class="card"><div class="card-h"><span class="h-title">' + esc(ph) + '</span></div><div class="card-b">' + (involved ? '<div class="mr-sub">' + esc(t("\u6D89\u53CA")) + "\uFF1A" + involved.split(",").map((x) => {
          const parts = x.trim().split("/");
          return parts.length >= 3 ? '<span class="ref-chip" data-go="obj/' + esc(parts[0]) + "/" + esc(parts[1]) + "/" + esc(parts[2]) + '">' + esc(x.trim()) + "</span>" : '<span class="small faint mono">' + esc(x.trim()) + "</span>";
        }).join(" ") + "</div>" : "") + '<div class="small faint">' + esc(t("\u4E8B\u4EF6\u94FE\uFF08\u5F52\u6863\u7A97\u53E3\uFF09\uFF1A")) + "</div>" + evRows + "</div></div>";
      });
    });
  }
  var RUN_CLS = { running: "acc", done: "ok", ok: "ok", success: "ok", failed: "esc", error: "esc", canceled: "mute", queued: "mute", pending: "mute" };
  var REG_CLS = { running: "acc", active: "acc", suspended: "warn", awaitingapproval: "warn", failed: "esc", error: "esc", done: "ok", ok: "ok", idle: "mute", registered: "mute" };
  var PH_CLS = { awaitingapproval: "warn", failed: "esc", done: "ok", executing: "acc", running: "acc" };
  function runChip(r) {
    const st = String(r && (r.status || r.state || r.phase) || t("\u672A\u77E5"));
    return chip(st, RUN_CLS[st] || "mute");
  }
  function phChip(ph) {
    const p = String(ph || "");
    return chip(p || "\u2014", PH_CLS[p.toLowerCase()] || "mute");
  }
  var WF_TEMPLATES = [
    { name: "bug \u751F\u547D\u5468\u671F", desc: "\u7F3A\u9677\u4ECE\u53D1\u73B0\u5230\u6536\u655B\u7684\u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\uFF08\u542B\u9A8C\u8BC1\u4E0E\u6279\u95F8\u4F4D\uFF09\u3002" },
    { name: "dev-delivery", desc: "\u5F00\u53D1\u4EA4\u4ED8\u6D41\u6C34\uFF08\u5B9E\u73B0\u2192\u8BC4\u5BA1\u2192\u843D\u5E93\uFF09\u7684\u76F8\u4F4D\u5E8F\u5217\u4E0E\u6279\u95F8\u4F4D\u3002" },
    { name: "\u6587\u6863\u6536\u655B", desc: "\u6587\u6863\u8349\u7A3F\u2192\u8BC4\u5BA1\u2192\u843D\u5E93\u7684\u6536\u655B\u76F8\u4F4D\u4E0E\u6279\u95F8\u4F4D\u3002" }
  ];
  function wfDesignRoute() {
    return "designdoc/" + encodeURIComponent("core_design") + "/" + encodeURIComponent("concepts/workflow.md");
  }
  function tplCardHtml(tp) {
    return '<div class="card"><div class="card-h"><span class="h-title">' + esc(tp.name) + '</span><span class="h-right">' + chip("Template", "mute") + '</span></div><div class="card-b"><div class="small">' + esc(t(tp.desc)) + '</div><div class="dh-actions" style="margin-top:8px">' + chipGo(t("\u8BBE\u8BA1\u89C1 workflow.md \xA79b"), wfDesignRoute(), "mute", true, t("\u8BBE\u8BA1\u6587\u6863\u5DE5\u4F5C\u53F0")) + ' <button class="btn sm" disabled title="' + esc(t("Template kind \u843D\u518C\u540E\u5F00\u653E\uFF08\xA79b \u56DB\u5C42 appliesTo\uFF09")) + '">' + esc(t("\u5B9E\u4F8B\u5316")) + "</button></div></div></div>";
  }
  function listScreen(el, ctx2) {
    const state = { rows: [], filter: "", err: "" };
    const snapRuns = ctx2.store.runs().slice(0, 8);
    const runRowHtml = (r) => {
      const wf = String(r && (r.workflow || r.name) || "");
      const no = r && (r.run != null ? r.run : r.no != null ? r.no : "");
      if (!wf && (no == null || no === "")) return "";
      const started = r && (r.startedAt || r.at || r.ts);
      const invChips = String(r && r.involved || "").split(",").map((x) => x.trim()).filter(Boolean).slice(0, 3).map((x) => {
        const rt = involvedRoute(x);
        return rt ? chipGo(trunc(x, 36), rt, "mute", true, t("\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0")) : '<span class="chip mute tiny">' + esc(trunc(x, 36)) + "</span>";
      }).join(" ");
      const go = wf && no != null && no !== "" ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(no)) : "";
      return "<tr" + (go ? ' data-go="' + esc(go) + '" style="cursor:pointer"' : "") + '><td class="mono small">#' + esc(String(no != null && no !== "" ? no : "\u2014")) + '</td><td class="mono small">' + esc(wf || "\u2014") + "</td><td>" + phChip(r && r.phase) + "</td><td>" + runChip(r) + '</td><td class="small mono">' + esc(relAge(started) || (started ? String(started).slice(0, 16) : "\u2014")) + '</td><td class="small">' + (invChips || '<span class="faint">\u2014</span>') + "</td></tr>";
    };
    const paintRuns = () => {
      const slot = el.querySelector("[data-wf-runs]");
      if (!slot) return;
      const seen = /* @__PURE__ */ new Set();
      const merged = [];
      for (const r of state.rows.concat(snapRuns)) {
        const k = String(r && (r.workflow || r.name) || "") + "#" + String(r && (r.run != null ? r.run : r.no != null ? r.no : ""));
        if (seen.has(k)) continue;
        seen.add(k);
        merged.push(r);
      }
      const shown = (state.filter ? merged.filter((r) => String(r && (r.workflow || r.name) || "") === state.filter) : merged).slice(0, 50);
      slot.innerHTML = '<div class="card"><div class="card-b" style="overflow-x:auto"><div style="margin-bottom:6px">' + chip(merged.length + t(" \u6761\uFF08runs \u9762\u5168\u7A97\u53E3+\u5FEB\u7167\u805A\u5408\uFF0C\u6309 workflow+run \u53BB\u91CD\uFF09"), "mute") + " " + (state.filter ? chipGo(t("\u6E05\u9664\u8FC7\u6EE4\uFF1A") + state.filter, "workflows", "warn", true, t("\u6E05\u9664\u8FC7\u6EE4\u5E76\u56DE\u5168\u8868")) : "") + "</div>" + (state.err ? '<div class="small faint" style="margin-bottom:6px">' + esc(t("\u8FD0\u884C\u8BB0\u5F55\u9762\u6682\u4E0D\u53EF\u8BFB\uFF08")) + esc(state.err) + esc(t("\uFF09\u2014\u2014\u4E0B\u8868\u4EC5\u5FEB\u7167\u805A\u5408\u884C\uFF08\u4E0D\u88C5\u5047\uFF09\u3002")) + "</div>" : "") + (shown.length ? '<table class="tbl"><thead><tr><th>#</th><th>' + esc(t("\u5DE5\u4F5C\u6D41")) + "</th><th>" + esc(t("\u76F8\u4F4D")) + "</th><th>" + esc(t("\u72B6\u6001")) + "</th><th>" + esc(t("\u5F00\u59CB")) + "</th><th>" + esc(t("\u6D89\u53CA")) + "</th></tr></thead><tbody>" + shown.map(runRowHtml).join("") + '</tbody></table><div class="small faint" style="margin-top:6px">' + esc(t("\u884C\u70B9\u51FB\u8FDB\u8FD0\u884C\u8BE6\u60C5\uFF1Binvolved chip=\u5BF9\u8C61\u5DE5\u4F5C\u53F0\u6DF1\u94FE\u3002")) + "</div>" : '<div class="small faint">' + esc(state.filter ? t("\u8BE5 workflow \u7A97\u53E3\u5185\u6682\u65E0\u8FD0\u884C\u884C\uFF08\u8BDA\u5B9E\u7A7A\u6001\uFF09\u3002") : t("\u7A97\u53E3\u5185\u6682\u65E0\u8FD0\u884C\u884C\uFF08runs \u9762\u4E0E\u5FEB\u7167\u805A\u5408\u5747\u7A7A\u2014\u2014\u8BDA\u5B9E\u7A7A\u6001\uFF09\u3002"))) + "</div></div>";
    };
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u5DE5\u4F5C\u6D41"),
      t("\u4E09\u533A=\u5728\u518C\u6CE8\u518C\u8868\uFF08GET /v1/workflows \u73B0\u8BFB\uFF09\u222A \u8FD0\u884C\u8BB0\u5F55\uFF08GET /v1/workflows/runs+\u5FEB\u7167\u805A\u5408\u5E76\u663E\uFF09\u222A \u6A21\u677F\uFF08\xA79b \u8BBE\u8BA1\u5361\uFF09\u3002"),
      chipGo(t("\u8BBE\u8BA1\u6587\u6863"), wfDesignRoute(), "mute", true, t("workflow.md \xA79b"))
    ) + '<h2 class="sec-title">' + esc(t("\u5728\u518C\u5DE5\u4F5C\u6D41")) + '</h2><div data-wf-reg><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u6CE8\u518C\u8868\u73B0\u8BFB\u4E2D\u2026")) + '</div></div></div></div><h2 class="sec-title" style="margin-top:18px">' + esc(t("\u8FD0\u884C\u8BB0\u5F55")) + '</h2><div data-wf-runs><div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u8FD0\u884C\u8BB0\u5F55\u73B0\u8BFB\u4E2D\u2026")) + '</div></div></div></div><h2 class="sec-title" style="margin-top:18px">' + esc(t("\u6A21\u677F")) + '</h2><div class="grid cols-2">' + WF_TEMPLATES.map(tplCardHtml).join("") + "</div></div>";
    paintRuns();
    (async () => {
      const stamp = typeof location !== "undefined" ? location.hash : "";
      const slot = el.querySelector("[data-wf-reg]");
      const r = await ctx2.api.workflows();
      if (!slot || !slot.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      if (!r.ok) {
        slot.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u6CE8\u518C\u8868\u6682\u4E0D\u53EF\u8BFB\uFF08")) + esc(String(r.error || r.status)) + esc(t("\uFF09\u2014\u2014\u4E0D\u88C5\u5047\uFF0C\u5237\u65B0\u91CD\u8BD5\u3002")) + "</div></div></div>";
        return;
      }
      const data = r.data && typeof r.data === "object" ? r.data : {};
      const rows = Array.isArray(data.workflows) ? data.workflows : [];
      if (!rows.length) {
        slot.innerHTML = '<div class="card"><div class="card-b"><div class="small faint">' + esc(t("\u6CE8\u518C\u8868\u4E3A\u7A7A\uFF08\u8BDA\u5B9E\u7A7A\u6001\u2014\u2014\u5F53\u524D\u65E0\u5728\u518C Workflow\uFF09\u3002")) + "</div></div></div>";
        return;
      }
      slot.innerHTML = '<div class="card"><div class="card-b" style="overflow-x:auto">' + chip((data.count != null ? data.count : rows.length) + t(" \u6761\u5728\u518C"), "acc") + " " + (data.scanCapped ? chip(t("\u626B\u63CF\u622A\u65AD\uFF08\u7A97\u53E3\u5185 ") + String(data.scanned != null ? data.scanned : "?") + t(" \u6761\uFF09"), "warn") : "") + '<table class="tbl" style="margin-top:8px"><thead><tr><th>' + esc(t("\u540D\u79F0")) + "</th><th>" + esc(t("\u72B6\u6001")) + "</th><th>" + esc(t("\u8FDE\u63A5")) + "</th><th>" + esc(t("\u6700\u8FD1")) + "</th><th></th></tr></thead><tbody>" + rows.map((w) => {
        const name = String(w && w.name || "");
        const ex = w && w.execution || {};
        const exState = String(ex.state || "");
        const connTxt = String(ex.connector || "") + (ex.connectorScope ? "/" + String(ex.connectorScope) : "");
        const latest = w && w.latest;
        const latestTxt = typeof latest === "string" ? latest : String(latest && (latest.summary || latest.note || latest.status || latest.state) || "");
        return '<tr><td class="mono small">' + (name ? '<span class="lk" data-go="wf/' + encodeURIComponent(name) + '">' + esc(name) + "</span>" : "\u2014") + "</td><td>" + chip(exState || "\u2014", REG_CLS[exState.toLowerCase()] || "mute") + '</td><td class="mono small">' + esc(connTxt || "\u2014") + '</td><td class="small">' + esc(trunc(latestTxt || "\u2014", 60)) + '</td><td style="text-align:right">' + (name ? '<span class="chip mute tiny click" data-wf-runfilter="' + esc(name) + '" title="' + esc(t("\u8FC7\u6EE4\u672C\u9875\u8FD0\u884C\u8BB0\u5F55")) + '">' + esc(t("\u8FD0\u884C\u8BB0\u5F55")) + "</span>" : "") + "</td></tr>";
      }).join("") + "</tbody></table></div></div>";
      slot.querySelectorAll("[data-wf-runfilter]").forEach((c) => {
        c.addEventListener("click", () => {
          state.filter = String(c.getAttribute("data-wf-runfilter") || "");
          paintRuns();
          const rs = el.querySelector("[data-wf-runs]");
          if (rs && rs.scrollIntoView) rs.scrollIntoView({ behavior: "smooth", block: "start" });
        });
      });
    })();
    (async () => {
      const stamp = typeof location !== "undefined" ? location.hash : "";
      const r = await ctx2.api.workflowRuns();
      if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      if (!r.ok) {
        state.err = String(r.error || r.status);
      } else {
        const data = r.data && typeof r.data === "object" ? r.data : {};
        state.rows = Array.isArray(data.runs) ? data.runs : [];
      }
      paintRuns();
    })();
  }
  function detailScreen(el, ctx2, name) {
    const wfs = ctx2.store.workflows();
    const runs = ctx2.store.runs().filter((r) => String(r && (r.workflow || r.name || "")) === String(name));
    const wf = wfs.find((w) => String(w && (w.name || w.workflow || "")) === String(name));
    const head = screenHead(t("\u5DE5\u4F5C\u6D41\u8BE6\u60C5"), "name=" + String(name), chipGo(t("\u2190 \u5DE5\u4F5C\u6D41"), "workflows", "mute", false));
    if (!wf && !runs.length) {
      el.innerHTML = '<div class="view-inner">' + head + emptyBox(t("\u5DE5\u4F5C\u6D41\u4E0D\u5728\u518C"), t("\u5FEB\u7167\u805A\u5408\u65E0\u6B64 workflow\uFF08\u53EF\u80FD\u5DF2\u4E0B\u7EBF\u6216\u540D\u6709\u8BEF\uFF09\u3002")) + "</div>";
      return;
    }
    const rows = [];
    if (wf) {
      if (wf.scope) rows.push(["scope", esc(String(wf.scope))]);
      if (wf.resourceVersion != null) rows.push(["rv", esc(String(wf.resourceVersion))]);
      if (wf.spec && wf.spec.title) rows.push([t("\u6807\u9898"), esc(String(wf.spec.title))]);
      if (wf.spec && wf.spec.goalRef) rows.push([t("\u76EE\u6807"), chipGo(String(wf.spec.goalRef), involvedRoute(String(wf.spec.goalRef)) || "docs", "mute", true, t("\u76EE\u6807\u6587\u6863"))]);
    }
    if (wf && wf.spec && wf.spec.dependsOn) rows.push([t("\u4F9D\u8D56"), (wf.spec.dependsOn || []).map((d) => chip(String(d), "mute")).join(" ")]);
    rows.push([t("\u8FD0\u884C"), chip(runs.length + t(" \u6B21"), runs.length ? "acc" : "mute")]);
    const runRows = runs.map((r, i) => {
      const no = r.no != null ? r.no : runs.length - i;
      return '<div class="mini-row" data-go="wf/' + encodeURIComponent(String(name)) + "/r/" + encodeURIComponent(String(no)) + '"><div class="mr-main"><div><b class="mono">run #' + esc(String(no)) + "</b> " + runChip(r) + '</div><div class="mr-sub small mono">' + esc(trunc(String(r.route || r.trigger || "\u2014"), 80)) + '</div></div><span class="w-ev-time mono">' + esc(relAge(r.startedAt || r.at || r.ts) || "\u2014") + "</span></div>";
    }).join("");
    const run0 = runs[0] || { workflow: String(name), phases: [], phase: "", status: "" };
    el.innerHTML = '<div class="view-inner">' + head + t('<div class="card"><div class="card-h"><span class="h-title">\u6982\u89C8</span></div><div class="card-b">') + kv(rows) + "</div></div>" + t('<h2 class="sec-title">\u8282\u70B9\u56FE</h2><div class="card"><div class="card-b" data-pg-box>') + phaseGraphHtml(run0, ctx2) + "</div></div>" + t('<h2 class="sec-title">\u8FD0\u884C\u8BB0\u5F55</h2>') + (runRows ? '<div class="card">' + runRows + "</div>" : t('<div class="small faint">\u8BE5 workflow \u6682\u65E0\u8FD0\u884C\u8BB0\u5F55\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>')) + "</div>";
    bindPhaseGraph(el, ctx2, run0);
  }
  function bindDocLinks(scopeEl, ctx2) {
    scopeEl.querySelectorAll("[data-go]").forEach((n) => {
      n.addEventListener("click", () => ctx2.go(n.getAttribute("data-go")));
    });
  }
  function runScreen(el, ctx2, name, no) {
    const runs = ctx2.store.runs().filter((r) => String(r && (r.workflow || r.name || "")) === String(name));
    const run = runs.find((r) => String(r.run != null ? r.run : r.no != null ? r.no : "") === String(no)) || runs[0];
    if (!run) {
      el.innerHTML = '<div class="view-inner"><div class="run-hero"><div class="rh-top">' + chipGo(t("\u2190 " + String(name)), "wf/" + encodeURIComponent(String(name)), "mute", false) + '</div><div class="rh-title">' + esc(String(name)) + ' <span class="rh-run">#' + esc(String(no)) + "</span></div></div>" + emptyBox(t("\u8FD0\u884C\u8BB0\u5F55\u4E0D\u5728\u518C"), t("\u5FEB\u7167 runs \u805A\u5408\u65E0\u6B64\u8FD0\u884C\uFF08\u53EF\u80FD\u5DF2\u51FA\u7A97\uFF09\u3002")) + "</div>";
      return;
    }
    const st = String(run.status || "");
    const stCls = st === "suspended" ? "warn" : RUN_CLS[st] || "mute";
    const startedAt = run.startedAt || run.at || run.ts;
    const gateWord = st === "suspended" ? t("\u7B49\u5BA1\u6279") : st === "running" ? t("\u8FD0\u884C\u4E2D") : st;
    const canGate = String(run.phase || "") === "AwaitingApproval" || st === "suspended";
    const gateAnchor = "Workflow/" + String(run.owner || "global") + "/" + String(run.workflow || name);
    el.innerHTML = '<div class="view-inner"><div class="run-hero"><div class="rh-top">' + chipGo(t("\u2190 " + String(name)), "wf/" + encodeURIComponent(String(name)), "mute", false) + '</div><div class="rh-title">' + esc(String(name)) + ' <span class="rh-run">#' + esc(String(run.run != null ? run.run : no)) + "</span>" + chip(st || t("\u672A\u77E5"), stCls) + '</div><div class="rh-meta">' + (startedAt ? "<span>" + esc(t("\u5F00\u59CB\u4E8E")) + " " + esc(relAge(startedAt) || String(startedAt)) + "</span>" : "") + (run.error ? '<span class="rh-err">' + esc(trunc(String(run.error), 120)) + "</span>" : "") + (st === "suspended" ? '<span class="ref-chip" data-go="approvals">' + esc(t("\u53BB\u5BA1\u6279 \u2192")) + "</span>" : "") + "</div>" + (canGate ? '<div class="dh-actions" style="margin-top:10px"><button class="btn sm primary" data-gate-act="grant">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn sm" data-gate-act="reject">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + '</button> <span class="small faint mono">' + esc(t("\u6279\u95F8\u951A\uFF1A")) + esc(gateAnchor) + "</span></div>" : "") + '</div><div class="step-line" data-step-line></div><div class="run-cols"><div class="run-col"><div class="panel-h">' + esc(t("\u5F53\u524D\u76F8\u4F4D")) + '</div><div data-run-now></div></div><div class="run-col"><div class="panel-h">' + esc(t("\u4E8B\u4EF6\u65F6\u95F4\u7EBF")) + '</div><div data-run-tl></div></div></div><div class="foot-note">' + esc(t("\u76F8\u4F4D\u5E8F\u5217=Workflow \u58F0\u660E\u539F\u6837\uFF1B\u63A8\u8FDB\u7531 patch \u589E\u91CF\u5B9E\u65F6\u5237\u65B0\uFF1B" + gateWord + "\u3002")) + "</div></div>";
    const sel = { idx: 0, phases: [] };
    const lineEl = el.querySelector("[data-step-line]");
    const paintLine = () => {
      const phases = Array.isArray(run.phases) && run.phases.length ? run.phases : [run.phase || t("\u5355\u8282\u70B9")];
      sel.phases = phases;
      const curIdx = phases.indexOf(run.phase);
      if (curIdx >= 0) sel.idx = curIdx;
      lineEl.innerHTML = phases.map((p, i) => {
        const done = st === "done" || (curIdx >= 0 ? i < curIdx : false);
        const active = i === curIdx && st !== "done";
        const dot = done ? "\u2713" : active ? "" : String(i + 1);
        const conn = i < phases.length - 1 ? '<div class="step-conn' + (done || curIdx >= 0 && i < curIdx ? " fill" : "") + '"></div>' : "";
        return '<div class="step' + (done ? " done" : "") + (active ? " active" : "") + (p === "Failed" ? " failstep" : "") + '" data-step="' + i + '" role="button"><div class="step-dot">' + dot + '</div><div class="step-name">' + esc(p) + "</div></div>" + conn;
      }).join("");
      el.querySelectorAll("[data-step]").forEach((n) => {
        n.addEventListener("click", () => {
          sel.idx = Number(n.getAttribute("data-step"));
          paintNow();
          paintSel();
        });
      });
      paintSel();
    };
    const paintSel = () => {
      el.querySelectorAll("[data-step]").forEach((n, i) => n.classList.toggle("sel", i === sel.idx));
    };
    const nowEl = el.querySelector("[data-run-now]");
    const tlEl = el.querySelector("[data-run-tl]");
    const paintNow = () => {
      const phases = sel.phases.length ? sel.phases : [run.phase || t("\u5355\u8282\u70B9")];
      const ph = phases[sel.idx] || run.phase || t("\u5355\u8282\u70B9");
      const isCur = ph === run.phase;
      const involved = String(run.involved || "");
      const refs = involved.split(",").map((x) => x.trim()).filter((x) => x.split("/").length >= 3).map((x) => {
        const p2 = x.split("/");
        return '<span class="ref-chip" data-go="obj/' + esc(p2[0]) + "/" + esc(p2[1]) + "/" + esc(p2[2]) + '">' + esc(x) + "</span>";
      }).join(" ");
      nowEl.innerHTML = '<div class="now-ph">' + esc(ph) + (isCur ? chip(gateWord || st || t("\u8FDB\u884C\u4E2D"), stCls) : "") + '</div><div class="small faint">' + esc(isCur ? t("\u6B63\u5728\u53D1\u751F\u2014\u2014\u63A8\u8FDB\u7531 patch \u5B9E\u65F6\u5237\u65B0") : t("\u5386\u53F2\u76F8\u4F4D\u2014\u2014\u4E8B\u4EF6\u6309\u5F52\u6863\u7A97\u53E3\u56DE\u653E")) + "</div>" + (refs ? '<div class="now-refs">' + esc(t("\u6D89\u53CA")) + "\uFF1A" + refs + "</div>" : "") + (isCur && st === "suspended" ? '<div class="now-act"><span class="ref-chip" data-go="approvals">' + esc(t("\u53BB\u5904\u7406\u8FD9\u6761\u5BA1\u6279")) + "</span></div>" : "");
      bindDocLinks(nowEl, ctx2);
    };
    const paintTl = () => {
      const trans = Array.isArray(run.transitions) ? run.transitions.filter(Boolean) : [];
      const transHtml = trans.length ? '<div class="small faint" style="margin-bottom:6px">' + esc(t("\u76F8\u4F4D\u8FC1\u79FB\uFF08runs \u9762\u73B0\u8BFB\uFF09\uFF1A")) + '</div><div class="tl" style="margin-bottom:10px">' + trans.slice(0, 20).map((x) => {
        const at = String(x && (x.at || x.ts || x.time) || "");
        const from = String(x && (x.from || x.fromPhase) || "\u2014");
        const to = String(x && (x.to || x.toPhase || x.phase) || "\u2014");
        return '<div class="tl-row"><div class="tl-dot"></div><div class="tl-main"><b>' + esc(from) + " \u2192 " + esc(to) + '</b></div><div class="tl-time mono">' + esc(relAge(at) || (at ? at.slice(0, 16) : "")) + "</div></div>";
      }).join("") + "</div>" : "";
      const evs = ctx2.store.events().filter((e) => {
        const hay = JSON.stringify([e.dedupKey, e.involved, e.type]);
        return hay.includes(String(run.workflow || "")) || hay.includes(String(run.phase || ""));
      }).slice(0, 12);
      tlEl.innerHTML = transHtml + (evs.length ? '<div class="tl">' + evs.map((e) => '<div class="tl-row"><div class="tl-dot"></div><div class="tl-main"><b>' + esc(String(e.type || "")) + '</b><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 56)) + '</div></div><div class="tl-time">' + esc(relAge(e.ts) || "") + "</div></div>").join("") + "</div>" : '<div class="tl-empty">' + esc(t("\u5F52\u6863\u7A97\u53E3\u5185\u6682\u65E0\u4E8B\u4EF6\u2014\u2014workflow \u63A8\u8FDB\u65F6\u6B64\u5904\u5B9E\u65F6\u843D\u884C\u3002")) + "</div>");
    };
    paintLine();
    paintNow();
    paintTl();
    const refreshFace = () => ctx2.api.workflowRuns().then((r) => {
      if (!r.ok || !el.isConnected) return;
      const arr = Array.isArray(r.data && r.data.runs) ? r.data.runs : Array.isArray(r.data) ? r.data : [];
      const full = arr.find((x) => String(x && (x.run != null ? x.run : "")) === String(no));
      if (!full) return;
      run.phases = Array.isArray(full.phases) ? full.phases : run.phases;
      run.phase = full.phase || run.phase;
      run.status = full.status || run.status;
      run.involved = full.involved || run.involved;
      run.transitions = Array.isArray(full.transitions) ? full.transitions : run.transitions;
      paintLine();
      paintNow();
      paintTl();
    });
    refreshFace();
    el.querySelectorAll("[data-gate-act]").forEach((btn) => {
      btn.addEventListener("click", async () => {
        const grant = btn.getAttribute("data-gate-act") === "grant";
        const confirmed = grant ? window.confirm(t("\u786E\u8BA4\u6279\u51C6\u7EE7\u7EED\uFF1F\u53D1\u5E03 ApprovalGranted\uFF08") + gateAnchor + t("\uFF09\u2014\u2014run \u6CBF\u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\u7EED\u884C\u3002")) : window.confirm(t("\u786E\u8BA4\u9A73\u56DE\u7EC8\u6B62\uFF1F\u53D1\u5E03 ApprovalRejected\uFF08") + gateAnchor + t("\uFF09\u2014\u2014run \u5C06\u8FDB\u5165\u7EC8\u6001 Failed\uFF08\u4E0D\u53EF\u6062\u590D\uFF09\u3002"));
        if (!confirmed) return;
        btn.disabled = true;
        btn.textContent = t("\u53D1\u5E03\u4E2D\u2026");
        const r = await ctx2.api.eventPub({ type: grant ? "ApprovalGranted" : "ApprovalRejected", involved: gateAnchor });
        if (!r.ok) {
          btn.disabled = false;
          btn.textContent = grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62");
          window.alert(t("\u4E8B\u4EF6\u53D1\u5E03\u5931\u8D25\uFF08") + String(r.error || r.status) + t("\uFF09\u2014\u2014run \u4ECD\u6302\u539F\u76F8\u4F4D\uFF0C\u672A\u88C5\u6210\u529F\u3002"));
          return;
        }
        refreshFace();
        ctx2.api.snapshot().then((s) => {
          if (s.ok) ctx2.store.init(s.data);
        });
      });
    });
  }
  function render3(el, ctx2) {
    const p = ctx2.params || {};
    if (p.n != null && p.n !== "") runScreen(el, ctx2, p.name, p.n);
    else if (p.name != null && p.name !== "") detailScreen(el, ctx2, p.name);
    else listScreen(el, ctx2);
  }

  // src/screens/events.js
  var filter = { level: "", q: "" };
  var LEVEL_CLS = { critical: "esc", error: "esc", warn: "warn", warning: "warn", info: "mute", debug: "mute" };
  function clean(x) {
    return String(x == null ? "" : x).replace(/^"+|"+$/g, "");
  }
  function levelChip(e) {
    const lv = String(e && e.level || "info");
    return chip(lv, LEVEL_CLS[lv] || "mute");
  }
  function involvedHtml(e) {
    const s = clean(e && e.involved);
    if (!s) return chip("involved\uFF1A\u2014", "mute");
    return s.split(/[,;]\s*/).filter(Boolean).map((part) => {
      const p = part.trim();
      const route = involvedRoute2(p);
      return route ? chipGo(p, route, "mute", true, t("\u67E5\u770B ") + p) : chip(p, "mute");
    }).join(" ");
  }
  function linksHtml(e) {
    const out = [];
    const route = String(e && e.route || "");
    const wf = wfAnchorOfRoute(route);
    if (wf) out.push(chipGo(t("workflow \u951A ") + wf, "wf/" + encodeURIComponent(wf), "acc", true, "route " + route + t(" \u2192 \u5DE5\u4F5C\u6D41")));
    if (e && e.run) out.push(chipGo("run #" + e.run, route && wf ? "wf/" + encodeURIComponent(wf) + "/r/" + encodeURIComponent(String(e.run)) : "workflows", "mute", true, t("\u5173\u8054\u8FD0\u884C")));
    const sess = clean(e && (e.agentSession || e.session));
    if (sess) out.push(chipGo(t("\u4F1A\u8BDD ") + trunc(sess, 18), "session/" + encodeURIComponent(sess), "mute", true, t("\u4F1A\u8BDD\u8BE6\u60C5")));
    const wake = e && (e.wake != null ? e.wake : e.loopId != null ? e.loopId : null);
    if (wake != null) out.push(chip(t("\u5FAA\u73AF\u5524\u9192 #") + wake, "mute"));
    return out.join(" ");
  }
  function eventRow(e) {
    const id = e && (e.id != null ? e.id : e.eventId != null ? e.eventId : "");
    const involved = clean(e && e.involved);
    const parts = involved.split(/[,;]\s*/).filter(Boolean);
    const invHtml = parts.length ? parts.map((part) => {
      const p = part.trim();
      const route = involvedRoute2(p);
      return route ? chipGo(p, route, "mute", true, t("\u67E5\u770B ") + p) : '<span class="mono">' + esc(p) + "</span>";
    }).join(" ") : '<span class="mono">\u2014</span>';
    const hasClipboard = typeof navigator !== "undefined" && !!(navigator.clipboard && navigator.clipboard.writeText);
    return '<div class="mini-row" data-go="event/' + encodeURIComponent(String(id)) + '"><span class="w-ev-dot" data-level="' + esc(String(e && e.level || "info")) + '"></span><div class="mr-main"><div>' + esc(String(e && e.type || t("\u672A\u547D\u540D\u4E8B\u4EF6"))) + " " + levelChip(e) + '</div><div class="mr-sub mono">' + invHtml + '</div></div><span class="w-ev-time mono">' + esc(relAge(e && (e.ts || e.at || e.time)) || "\u2014") + '</span><button class="btn sm" data-copy-id="' + esc(String(id)) + '"' + (hasClipboard ? "" : ' disabled title="' + esc(t("\u526A\u8D34\u677F\u4E0D\u53EF\u7528")) + '"') + ">" + esc(t("\u590D\u5236 id")) + "</button></div>";
  }
  function listScreen2(el, ctx2) {
    let rows = ctx2.store.events().slice();
    if (filter.level) rows = rows.filter((e) => String(e && e.level || "") === filter.level);
    if (filter.q) {
      const q = filter.q.toLowerCase();
      rows = rows.filter((e) => String(e && (e.type || "") + " " + (e.involved || "") + " " + (e.summary || e.message || "")).toLowerCase().includes(q));
    }
    const fbar = '<div class="filter-bar">' + t('<input class="inp" id="ev-q" placeholder="\u6309\u7C7B\u578B/\u5BF9\u8C61/\u6458\u8981\u8FC7\u6EE4\u2026" value="') + esc(filter.q) + '">' + ["", "error", "warn", "info"].map(
      (lv) => '<span class="fchip' + (filter.level === lv ? " on" : "") + '" data-act="ev-level" data-lv="' + esc(lv) + '">' + esc(lv || t("\u5168\u90E8\u7EA7\u522B")) + "</span>"
    ).join("") + "</div>";
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u4E8B\u4EF6\u6D41"), t("\u5F52\u6863\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408 + patch \u589E\u91CF\uFF1B\u70B9\u51FB\u770B\u5168\u6587\u4E0E\u94FE\u8DEF\uFF09")) + fbar + (rows.length ? '<div class="card">' + rows.slice(0, 300).map(eventRow).join("") + (rows.length > 300 ? t('<div class="small faint" style="padding:8px 14px">\u2026\u5171 ') + rows.length + t(" \u884C\u2014\u2014\u7528\u8FC7\u6EE4\u6536\u7A84\u3002</div>") : "") + "</div>" : emptyBox(t("\u7A97\u53E3\u5185\u65E0\u4E8B\u4EF6"), filter.q || filter.level ? t("\u5F53\u524D\u8FC7\u6EE4\u96F6\u547D\u4E2D\u2014\u2014\u653E\u5BBD\u6761\u4EF6\u518D\u8BD5\u3002") : t("\u5FEB\u7167 events \u805A\u5408\u4E3A\u7A7A\uFF08\u4E0D\u88C5\u5047\uFF09\u3002"))) + "</div>";
    const inp = el.querySelector("#ev-q");
    if (inp) {
      inp.addEventListener("input", () => {
        filter.q = inp.value;
        listScreen2(el, ctx2);
      });
      inp.addEventListener("keydown", (ev) => {
        if (ev.key === "Enter") ev.preventDefault();
      });
    }
    el.querySelectorAll('[data-act="ev-level"]').forEach((c) => {
      c.addEventListener("click", () => {
        filter.level = c.dataset.lv || "";
        listScreen2(el, ctx2);
      });
    });
    el.querySelectorAll("[data-copy-id]").forEach((btn) => {
      if (btn.disabled) return;
      btn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        const raw = String(btn.dataset.copyId || "");
        navigator.clipboard.writeText(raw).then(() => {
          btn.textContent = t("\u5DF2\u590D\u5236");
          setTimeout(() => {
            btn.textContent = t("\u590D\u5236 id");
          }, 1200);
        }).catch(() => {
          btn.textContent = t("\u590D\u5236\u5931\u8D25");
          setTimeout(() => {
            btn.textContent = t("\u590D\u5236 id");
          }, 1200);
        });
      });
    });
  }
  async function detailScreen2(el, ctx2, id) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u4E8B\u4EF6\u8BE6\u60C5"), "id=" + id) + '<div class="card"><div class="card-b">' + kv([[t("\u72B6\u6001"), chip(t("\u73B0\u8BFB\u4E2D\u2026"), "mute")]]) + '</div></div><div class="dh-actions" style="margin-top:10px">' + chipGo(t("\u2190 \u56DE\u4E8B\u4EF6\u6D41"), "events", "mute", false) + "</div></div>";
    const r = await ctx2.api.event(id);
    if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    const box = el.querySelector(".card .card-b");
    if (!box) return;
    if (!r.ok) {
      box.innerHTML = t('<div class="ob-empty">\u4E8B\u4EF6\u5168\u6587\u6682\u4E0D\u53EF\u8FBE\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u6062\u590D\u540E\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>");
      return;
    }
    const e = r.data && typeof r.data === "object" ? r.data : { raw: r.data };
    const known = ["type", "level", "ts", "at", "time", "involved", "dedupKey", "routedTo", "summary", "message", "route", "run", "agentSession", "session", "wake"];
    const rows = known.filter((k) => e[k] != null && e[k] !== "").map((k) => {
      const v = k === "involved" || k === "summary" || k === "message" || k === "agentSession" || k === "session" ? clean(e[k]) : String(e[k]);
      return [k, k === "ts" || k === "at" || k === "time" ? timeHtml(e[k]) : esc(trunc(v, 200))];
    });
    box.innerHTML = '<div class="ev-head">' + esc(String(e.type || t("\u672A\u547D\u540D\u4E8B\u4EF6"))) + " " + levelChip(e) + " " + linksHtml(e) + "</div>" + (rows.length ? kv(rows) : "") + (e.involved ? '<div class="ev-involved">' + involvedHtml(e) + "</div>" : "") + t('<details class="md-raw"><summary>\u539F\u59CB\u8F7D\u8377\uFF08JSON\uFF09</summary><pre class="mono md-raw-pre">') + esc(JSON.stringify(e, null, 2)) + "</pre></details>";
  }
  function render4(el, ctx2) {
    const id = ctx2.params && ctx2.params.id;
    if (id != null && id !== "") detailScreen2(el, ctx2, String(id));
    else listScreen2(el, ctx2);
  }

  // src/screens/objects.js
  function tagGroups(tags, vocab) {
    const V = vocab || {};
    const m = {};
    (tags || []).forEach((t2) => {
      const type2 = V[t2] || t("\u672A\u5206\u7EC4");
      (m[type2] = m[type2] || []).push(t2);
    });
    return Object.keys(m).sort((a, b) => {
      if (a === t("\u672A\u5206\u7EC4")) return 1;
      if (b === t("\u672A\u5206\u7EC4")) return -1;
      return a < b ? -1 : 1;
    }).map((type2) => ({ type: type2, tags: m[type2].sort() }));
  }
  function rulesFor(obj, rules) {
    if (!rules) return { hits: [], degraded: true, reason: t("\u89C4\u5219\u6587\u6863\u65CF\u7D22\u5F15\u6682\u4E0D\u53EF\u8FBE\u2014\u2014\u547D\u4E2D\u5224\u5B9A\u6682\u7F3A\uFF08\u4E0D\u5192\u5145\u65E0\u9002\u7528\uFF09\u3002"), scanned: 0 };
    const o = obj || {};
    const text2 = String(o.content || "");
    const low = text2.toLowerCase();
    const nameLow = String(o.name || "").toLowerCase();
    const scopeLow = String(o.scope || "").toLowerCase();
    const kindLow = String(o.kind || "").toLowerCase();
    const objTags = o.tags || [];
    const out = [];
    (rules || []).forEach((r) => {
      if (!r || !r.path) return;
      const base = String(r.path).replace(/^.*\//, "").replace(/\.md$/, "");
      const rlow = String(r.content || "").toLowerCase();
      const rtags = r.tags || [];
      const why = [];
      const tagHit = objTags.filter((t2) => rtags.includes(t2));
      if (tagHit.length) why.push(t("\u6807\u7B7E\u300C") + tagHit.join("\u3001") + t("\u300D\u540C\u7EC4"));
      if (kindLow && rlow.includes(kindLow)) why.push(t("\u89C4\u5219\u539F\u6587\u63D0\u53CA ") + o.kind);
      if (scopeLow && scopeLow !== "global" && scopeLow !== "builtin" && rlow.includes(scopeLow)) why.push(t("\u89C4\u5219\u539F\u6587\u63D0\u53CA ") + o.scope);
      const nameStem = nameLow.replace(/[-_].*$/, "");
      if (nameStem && nameStem.length >= 3 && rlow.includes(nameStem)) why.push(t("\u89C4\u5219\u539F\u6587\u63D0\u53CA ") + nameStem);
      if (base && low && low.includes(base.toLowerCase())) why.push(t("\u5BF9\u8C61\u6B63\u6587\u5F15\u7528 ") + base);
      if (!why.length) return;
      out.push({ ref: (r.repo || "core_design") + "/" + r.path + "#" + base, name: base, why });
    });
    out.sort((a, b) => b.why.length - a.why.length || (a.name < b.name ? -1 : 1));
    return { hits: out.slice(0, 8), degraded: false, reason: "", scanned: (rules || []).length };
  }
  function sectionCompletion(content, requiredSections) {
    const secs = requiredSections || [];
    if (!secs.length) return [];
    const heads = {};
    String(content || "").split("\n").forEach((line) => {
      const m = /^\s*#{1,6}\s+(.+?)\s*$/.exec(line);
      if (m) heads[m[1].toLowerCase()] = true;
    });
    return secs.map((s) => {
      const key = String(s).toLowerCase();
      return { name: String(s), done: !!heads[key] || String(content || "").toLowerCase().includes(key) };
    });
  }
  function tagsOfDoc(content) {
    const t2 = loadYaml(content);
    if (!t2) return [];
    const spec = t2.spec && typeof t2.spec === "object" ? t2.spec : {};
    const meta = spec.metadata && typeof spec.metadata === "object" ? spec.metadata : t2.metadata && typeof t2.metadata === "object" ? t2.metadata : {};
    const tags = meta.tags || spec.tags || t2.tags;
    return Array.isArray(tags) ? tags.map((x) => String(x)) : [];
  }
  function isMarkdownDoc(kind, name) {
    if (kind === "Goal" || kind === "Rule" || kind === "Runbook" || kind === "DriftReport") return true;
    return /(^|[-_])(README|design|report|notes)([-_.]|$)/i.test(String(name || ""));
  }
  function tagPanel(groups) {
    if (!groups || !groups.length) {
      return t('<div class="ob-empty">\u6682\u65E0\u6807\u7B7E\u2014\u2014\u5BF9\u8C61\u6B63\u6587\u672A\u58F0\u660E spec.metadata.tags\uFF08\u58F0\u660E\u7ECF Apply \u843D\u5E93\u540E\u6B64\u5904\u73B0\u8BFB\uFF09\u3002</div>');
    }
    return groups.map(
      (g) => '<div class="ob-tagrow">' + chip(g.type, "mute") + " " + g.tags.map((t2) => '<span class="tag click" data-obj-tag="' + esc(t2) + t('" title="\u770B\u540C tag \u5BF9\u8C61">#') + esc(t2) + "</span>").join(" ") + "</div>"
    ).join("");
  }
  function rulePanel(res) {
    if (res.degraded) return '<div class="ob-empty">' + esc(res.reason) + "</div>";
    if (!res.hits.length) {
      return t('<div class="ob-empty">\u65E0\u9002\u7528\u89C4\u5219\u2014\u2014\u6309 kind/scope/tag/\u6B63\u6587\u5F15\u7528\u56DB\u8DEF\u68C0\u67E5\u96F6\u547D\u4E2D\uFF08\u68C0\u67E5\u8303\u56F4=') + res.scanned + t(" \u4EFD\u89C4\u5219\u6587\u6863\uFF09\u3002</div>");
    }
    return res.hits.map(
      (h, i) => '<details class="ob-rule"><summary><span class="lk mono" data-go="' + esc("designdoc/core_design/" + encodeURIComponent(String(h.ref.replace(/^[^/]+\/([^#]+).*/, "$1") + ".md"))) + '">' + esc(h.ref) + '</span></summary><div class="ob-rule-why small">' + h.why.map((w) => chip(w, "mute")).join(" ") + "</div></details>"
    ).join("");
  }
  function ctxPanel(rv) {
    return '<div class="ob-ctx">' + chip(t("\u5DF2\u5165\u5FEB\u7167"), "ok") + ' <span class="small">resourceVersion r' + esc(String(rv || 0)) + t("\uFF08rv \u672A\u53D8\u4E0D\u91CD\u8BFB\u6B63\u6587\u2014\u2014\u964D\u7EA7\u8F6E\u8BE2\u8DEF\u5F84\u540C\u6837\u751F\u6548\uFF09</span>") + " " + chipGo(t("\u5728\u6587\u6863\u5E93\u4E2D\u6253\u5F00"), "docs", "mute", true) + "</div>";
  }
  function opsPanel(ref) {
    return '<div><button class="btn sm" data-go="' + esc(docRoute(ref.kind, ref.scope, ref.name)) + t('">\u7F16\u8F91\u5668</button> ') + '<button class="btn sm" data-go="' + esc("s/" + encodeURIComponent(ref.scope)) + t('">\u6240\u5C5E scope</button> ') + "</div>";
  }
  function schemePanel(kindReg, content) {
    if (!kindReg) return t('<div class="ob-empty">scheme \u6682\u4E0D\u53EF\u8FBE\uFF08kinds \u805A\u5408\u7F3A\u5E2D\uFF09\u3002</div>');
    const spec = kindReg.spec || {};
    const keys = Object.keys(spec);
    const cards = keys.length ? yamlToCards(Object.keys(spec).map((f) => f + ": " + JSON.stringify(spec[f])).join("\n")) : "";
    const comp = sectionCompletion(content, kindReg.requiredSections || []).map(
      (c) => chip(c.name + (c.done ? " \u2713" : t(" \u7F3A")), c.done ? "ok" : "mute")
    ).join(" ");
    if (!cards && !comp) return "";
    return (comp ? '<div class="ob-seccomp">' + comp + "</div>" : "") + (cards ? '<div data-md-cards="1">' + cards + "</div>" : t('<div class="ob-empty">\u8BE5 kind \u65E0\u5B57\u6BB5\u58F0\u660E\uFF08spec \u7F3A\u5E2D\uFF09\u3002</div>'));
  }
  function contentPanel(kind, name, content) {
    if (!String(content || "").trim()) return t('<div class="ob-empty">\u672C\u6587\u6863\u6682\u65E0\u6B63\u6587\u3002</div>');
    if (isMarkdownDoc(kind, name)) {
      return '<div data-md-render="1">' + renderMarkdown(content) + "</div>" + rawDisclosure(t("\u539F\u6587"), content);
    }
    return '<div data-md-cards="1">' + (yamlToCards(content) || t('<div class="ob-empty">\uFF08\u7A7A\u58F0\u660E\uFF09</div>')) + "</div>" + rawDisclosure("YAML", content);
  }
  async function hydrateDoc(el, ctx2, ref) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const cached = ctx2.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
    if (cached) return cached.content;
    const r = await ctx2.api.object(ref.kind, ref.scope, ref.name);
    if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return null;
    if (!r.ok) return null;
    const data = r.data && typeof r.data === "object" ? r.data : {};
    const content = data.content != null ? String(data.content) : typeof r.data === "string" ? r.data : JSON.stringify(r.data, null, 2);
    ctx2.store.cacheDoc(ref.kind, ref.scope, ref.name, ref.rv, content);
    return content;
  }
  async function hydrateRules(el, ctx2) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const ruleRefs = ctx2.store.refs("Rule").slice(0, 12);
    if (!ruleRefs.length) return null;
    const rules = [];
    for (const rr of ruleRefs) {
      let content = "";
      const cached = ctx2.store.cachedDoc(rr.kind, rr.scope, rr.name, rr.rv);
      if (cached) content = cached.content;
      else {
        const r = await ctx2.api.object(rr.kind, rr.scope, rr.name);
        if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return null;
        if (r.ok) {
          const data = r.data && typeof r.data === "object" ? r.data : {};
          content = data.content != null ? String(data.content) : "";
          if (content) ctx2.store.cacheDoc(rr.kind, rr.scope, rr.name, rr.rv, content);
        }
      }
      rules.push({ repo: rr.scope, path: rr.name + ".md", tags: tagsOfDoc(content), content });
    }
    return rules;
  }
  function renderScreen(el, ctx2, ref) {
    const kinds = ctx2.store.kinds();
    const kindReg = kinds.find((k) => k && k.name === ref.kind) || null;
    const cached = ctx2.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
    const content = cached ? cached.content : "";
    const tags = tagsOfDoc(content);
    const ruleState = rulesPending.get(ref.kind + "|" + ref.name) || null;
    el.innerHTML = '<div class="view-inner">' + screenHead(
      ref.name,
      t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0\uFF08\u516D\u7247\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme\uFF09"),
      kindBadge(ref.kind) + " " + chip("r" + String(ref.rv || 0), "mute")
    ) + '<div class="obj-grid">' + t('<section class="card"><div class="card-h"><span class="h-title">\u6807\u7B7E</span></div><div class="card-b" data-panel="tags">') + tagPanel(tagGroups(tags)) + "</div></section>" + t('<section class="card"><div class="card-h"><span class="h-title">\u9002\u7528\u89C4\u5219</span></div><div class="card-b" data-panel="rules">') + (ruleState ? rulePanel(ruleState) : t('<div class="small faint">\u89C4\u5219\u547D\u4E2D\u5224\u5B9A\u4E2D\u2026</div>')) + "</div></section>" + t('<section class="card"><div class="card-h"><span class="h-title">\u4E0A\u4E0B\u6587</span></div><div class="card-b">') + ctxPanel(ref.rv) + "</div></section>" + t('<section class="card"><div class="card-h"><span class="h-title">\u64CD\u4F5C</span></div><div class="card-b">') + opsPanel(ref) + "</div></section>" + t('<section class="card obj-content"><div class="card-h"><span class="h-title">\u5185\u5BB9</span><span class="h-right">') + chip("scope: " + ref.scope, "mute") + '</span></div><div class="card-b" data-panel="content">' + (cached ? contentPanel(ref.kind, ref.name, content) : t('<div class="small faint">\u6B63\u6587\u73B0\u8BFB\u4E2D\u2026</div>')) + '</div></section><section class="card"><div class="card-h"><span class="h-title">scheme</span></div><div class="card-b">' + schemePanel(kindReg, content) + "</div></section></div></div>";
    el.querySelectorAll("[data-obj-tag]").forEach((t2) => {
      t2.addEventListener("click", () => ctx2.go("docs"));
    });
    if (!cached) {
      hydrateDoc(el, ctx2, ref).then((c) => {
        if (c == null || !el.isConnected) return;
        renderScreen(el, ctx2, ref);
      });
    }
    if (!ruleState) {
      hydrateRules(el, ctx2).then((rules) => {
        if (rules == null || !el.isConnected) return;
        const fresh = ctx2.store.cachedDoc(ref.kind, ref.scope, ref.name, ref.rv);
        const res = rulesFor({ kind: ref.kind, scope: ref.scope, name: ref.name, content: fresh ? fresh.content : "", tags }, rules);
        rulesPending.set(ref.kind + "|" + ref.name, res);
        const panel = el.querySelector('[data-panel="rules"]');
        if (panel) panel.innerHTML = rulePanel(res);
      });
    }
  }
  var rulesPending = /* @__PURE__ */ new Map();
  function render5(el, ctx2) {
    const p = ctx2.params || {};
    let objs = ctx2.store.objects();
    let ref = null;
    if (p.kind && p.name) {
      ref = objs.find((o) => o.kind === p.kind && o.name === p.name && (!p.scope || p.scope === "global" || o.scope === p.scope)) || null;
    } else if (p.kind) {
      ref = objs.find((o) => o.kind === p.kind) || null;
    } else {
      ref = objs[0] || null;
    }
    if (!ref) {
      const hint = p.kind ? "kind\u300C" + p.kind + "\u300D" : t("\u5FEB\u7167");
      el.innerHTML = '<div class="view-inner">' + screenHead(t("\u5BF9\u8C61\u5DE5\u4F5C\u53F0"), t("\u516D\u7247\u56FA\u5B9A\uFF1A\u6807\u7B7E/\u89C4\u5219/\u4E0A\u4E0B\u6587/\u64CD\u4F5C/\u5185\u5BB9/scheme")) + emptyBox(hint + t("\u6682\u65E0\u5728\u518C\u5BF9\u8C61"), t("\u5FEB\u7167 objects \u805A\u5408\u4E3A\u7A7A\u6216\u8BE5 kind \u65E0\u5BF9\u8C61\uFF08\u4E0D\u88C5\u5047\uFF09\u3002")) + "</div>";
      return;
    }
    renderScreen(el, ctx2, ref);
  }

  // src/screens/tasks.js
  var PHASE_VIEW = {
    running: { label: t("\u6B63\u5728\u505A"), chip: "acc" },
    verifying: { label: t("\u9A8C\u8BC1\u4E2D"), chip: "acc" },
    queued: { label: t("\u6392\u961F\u4E2D"), chip: "mute" },
    ready: { label: t("\u5F85\u542F\u52A8"), chip: "mute" },
    blocked: { label: t("\u53D7\u963B"), chip: "warn" },
    failed: { label: t("\u5931\u8D25"), chip: "esc" },
    done: { label: t("\u5DF2\u5B8C\u6210"), chip: "ok" },
    canceled: { label: t("\u5DF2\u53D6\u6D88"), chip: "mute" }
  };
  function statusOf(t2) {
    return t2 && t2.status || {};
  }
  function specOf(t2) {
    return t2 && t2.spec || {};
  }
  function phaseChip(t2) {
    const phase = String(statusOf(t2).phase || "");
    const v = PHASE_VIEW[phase];
    if (!v) return chip(phase || t("\u72B6\u6001\u672A\u77E5"), "mute");
    return chip(v.label, v.chip);
  }
  function humanWait(t2) {
    const s = statusOf(t2);
    return /审批|批准|待审|等待人|approval|approve/i.test(String(s.lastError || "") + " " + String(s.lastAction || ""));
  }
  function groupOf(t2) {
    const phase = String(statusOf(t2).phase || "");
    if (phase === "running" || phase === "verifying") return "doing";
    if (phase === "done" || phase === "canceled") return "done";
    if (phase === "blocked" || phase === "failed") return humanWait(t2) ? "waiting" : "stuck";
    return "queued";
  }
  var GROUPS = [
    { id: "doing", label: t("\u6B63\u5728\u505A") },
    { id: "waiting", label: t("\u7B49\u4EBA") },
    { id: "stuck", label: t("\u5361\u4F4F") },
    { id: "done", label: t("\u5B8C\u6210") },
    { id: "queued", label: t("\u6392\u961F") }
  ];
  function groupTasks(tasks) {
    const out = { doing: [], waiting: [], stuck: [], done: [], queued: [] };
    (tasks || []).forEach((t2) => {
      out[groupOf(t2)].push(t2);
    });
    return out;
  }
  function taskIdOf(t2) {
    if (!t2) return "";
    if (t2.taskId) return String(t2.taskId);
    return (t2.scope || "global") + "/" + String(t2.name || "");
  }
  function taskTitle(t2) {
    const spec = specOf(t2);
    return String(spec.title || t2 && t2.title || t2 && t2.name || t("\u672A\u547D\u540D\u4EFB\u52A1"));
  }
  var cancelReq = /* @__PURE__ */ new Set();
  function cancellable(t2) {
    const phase = String(statusOf(t2).phase || "");
    return !phase || phase === "queued" || phase === "pending";
  }
  function cancelBtnHtml(t2) {
    const id = taskIdOf(t2);
    if (cancelReq.has(id)) return chip(t("\u53D6\u6D88\u8BF7\u6C42\u5DF2\u53D1\u5E03"), "warn");
    if (!cancellable(t2)) return "";
    return '<button class="btn sm" data-task-cancel="' + esc(id) + '">' + esc(t("\u53D6\u6D88")) + "</button>";
  }
  function requestCancel(el, ctx2, t2) {
    const name = String(t2 && t2.name || "");
    const scope = String(t2 && t2.scope || "global") || "global";
    if (!name) {
      alert(t("\u4EFB\u52A1\u7F3A name\u2014\u2014\u65E0\u6CD5\u6784\u9020 Task/<scope>/<name> \u8BCD\u5F62\uFF0C\u53D6\u6D88\u672A\u53D1\u5E03"));
      return;
    }
    const ok = confirm(t("\u53D6\u6D88\u4EFB\u52A1 ") + name + t("\uFF1F\u53D1\u5E03 TaskCancelRequested\uFF08\u8C03\u5EA6\u5668\u5F52\u6863\u9762\u6D88\u8D39\u2014\u2014\u4E0D\u4FDD\u8BC1\u5DF2\u5728\u8DD1\u7684\u8FDB\u7A0B\u7ACB\u523B\u7EC8\u6B62\uFF09"));
    if (!ok) return;
    ctx2.api.eventPub({
      type: "TaskCancelRequested",
      involved: "Task/" + scope + "/" + name,
      dedupKey: "console-cancel-" + name + "-" + Date.now()
    }).then((r) => {
      if (r && r.ok) {
        cancelReq.add(taskIdOf(t2));
        ctx2.api.snapshot().then((s) => {
          if (s && s.ok) ctx2.store.init(s.data);
        }).catch(() => {
        });
        if (el.isConnected) render6(el, ctx2);
      } else {
        alert(t("\u53D6\u6D88\u8BF7\u6C42\u53D1\u5E03\u5931\u8D25\uFF1A") + String(r && (r.error || r.status) || t("\u7F51\u7EDC\u5F02\u5E38")));
      }
    }).catch((err) => {
      alert(t("\u53D6\u6D88\u8BF7\u6C42\u53D1\u5E03\u5931\u8D25\uFF1A") + String(err && err.message || err || t("\u7F51\u7EDC\u5F02\u5E38")));
    });
  }
  function taskRow(t2) {
    const id = taskIdOf(t2);
    const s = statusOf(t2);
    const spec = specOf(t2);
    const deps = (spec.dependsOn || s.dependsOn || []).length;
    const agent = String(s.agent || spec.agent || "").trim();
    const sess = String(s.agentSession || s.sessionId || "").trim();
    const last = String(s.lastAction || "").trim();
    const unknown = !String(s.phase || "");
    return '<div class="task-row"><div class="task-main"><div class="task-title"><b>' + esc(taskTitle(t2)) + "</b>" + phaseChip(t2) + chip(id, "mute") + (deps ? chip(t("\u4F9D\u8D56 ") + deps, "mute") : "") + (agent ? chip(t("\u8D1F\u8D23 ") + agent, "mute") : "") + '</div><div class="task-sub small">' + esc(last || t("\u6700\u8FD1\u52A8\u4F5C\uFF1A\u2014")) + (String(s.lastError || "").trim() ? " " + chip(trunc(String(s.lastError), 60), "esc") : "") + "</div>" + (unknown ? '<div class="task-sub small faint">' + esc(t("\u8C03\u5EA6\u5668\u672A\u89C1 status\u2014\u2014\u5DF2\u77E5\u6D3B\u4F53 bug\uFF08BUG \u53F0\u8D26\uFF1A\u7A7A status \u4E0D\u8C03\u5EA6\uFF09\uFF1B\u53EF\u7528\u53D6\u6D88\u52A8\u4F5C\u9003\u751F")) + "</div>" : "") + '</div><div class="task-acts">' + (sess ? chipGo(t("\u5BF9\u8BDD"), "session/" + encodeURIComponent(sess), "mute", true, t("\u5173\u8054\u4F1A\u8BDD")) : "") + cancelBtnHtml(t2) + "</div></div>";
  }
  function tasksScreen(el, ctx2) {
    const tasks = ctx2.store.tasks();
    const g = groupTasks(tasks);
    const body = GROUPS.map((grp) => {
      const rows = g[grp.id] || [];
      return '<div class="card"><div class="card-h"><span class="h-title">' + esc(grp.label) + "</span>" + chip(String(rows.length) + t(" \u9879"), rows.length ? "acc" : "mute") + "</div>" + (rows.length ? '<div class="card-b">' + rows.map(taskRow).join("") + "</div>" : '<div class="card-b small faint">' + esc(grp.label) + t("\uFF1A\u6682\u65E0\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>")) + "</div>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u4EFB\u52A1\u4E2D\u5FC3"), t("\u5206\u7EC4=\u786E\u5B9A\u6027\u6295\u5F71\uFF08\u72B6\u6001\u5355\u4E00\u5199\u8005=\u8C03\u5EA6\u5668\uFF09\uFF1B\u70B9\u51FB\u884C\u8FDB\u8BE6\u60C5"), chipGo(t("\u5173\u6CE8\u62A5\u544A"), "report", "acc", false)) + '<div class="grid cols-2">' + body + "</div></div>";
    el.querySelectorAll(".task-row").forEach((row, i) => {
      let n = 0;
      for (const grp of GROUPS) {
        const rows = g[grp.id] || [];
        if (i < n + rows.length) {
          const t2 = rows[i - n];
          row.style.cursor = "pointer";
          row.addEventListener("click", () => ctx2.go("task/" + encodeURIComponent(taskIdOf(t2))));
          return;
        }
        n += rows.length;
      }
    });
    const byId = new Map(tasks.map((x) => [taskIdOf(x), x]));
    el.querySelectorAll("[data-task-cancel]").forEach((btn) => {
      btn.addEventListener("click", (ev) => {
        ev.stopPropagation();
        const t2 = byId.get(btn.dataset.taskCancel);
        if (t2) requestCancel(el, ctx2, t2);
      });
    });
  }
  function taskDetailScreen(el, ctx2, id) {
    const tasks = ctx2.store.tasks();
    const t2 = tasks.find((x) => taskIdOf(x) === String(id));
    const head = screenHead(t("\u4EFB\u52A1\u8BE6\u60C5"), "id=" + String(id), chipGo(t("\u2190 \u4EFB\u52A1\u4E2D\u5FC3"), "tasks", "mute", false));
    if (!t2) {
      el.innerHTML = '<div class="view-inner">' + head + emptyBox(t("\u4EFB\u52A1\u4E0D\u5728\u518C"), t("\u5FEB\u7167\u4EFB\u52A1\u6E05\u5355\u65E0\u6B64 id\uFF08\u53EF\u80FD\u5DF2\u5B8C\u6210\u51FA\u7A97\u6216 id \u6709\u8BEF\uFF09\u3002")) + "</div>";
      return;
    }
    const s = statusOf(t2);
    const spec = specOf(t2);
    const rows = [
      [t("\u76F8\u4F4D"), phaseChip(t2)],
      [t("\u6807\u9898"), esc(taskTitle(t2))],
      ["id", esc(taskIdOf(t2))]
    ];
    if (spec.agent || s.agent) rows.push([t("\u8D1F\u8D23"), esc(String(s.agent || spec.agent))]);
    const sessId = String(s.agentSession || s.sessionId || "").trim();
    if (sessId) rows.push([t("\u4F1A\u8BDD"), chipGo(trunc(sessId, 24), "session/" + encodeURIComponent(sessId), "acc", true, t("\u5173\u8054\u4F1A\u8BDD"))]);
    if (s.lastAction) rows.push([t("\u6700\u8FD1\u52A8\u4F5C"), esc(trunc(String(s.lastAction), 200))]);
    if (s.lastError) rows.push([t("\u6700\u8FD1\u9519\u8BEF"), chip(trunc(String(s.lastError), 120), "esc")]);
    if (!String(s.phase || "")) rows.push([t("\u6CE8\u8BB0"), esc(t("\u8C03\u5EA6\u5668\u672A\u89C1 status\u2014\u2014\u5DF2\u77E5\u6D3B\u4F53 bug\uFF08BUG \u53F0\u8D26\uFF1A\u7A7A status \u4E0D\u8C03\u5EA6\uFF09\uFF1B\u53EF\u7528\u53D6\u6D88\u52A8\u4F5C\u9003\u751F"))]);
    const deps = spec.dependsOn || s.dependsOn || [];
    if (deps.length) rows.push([t("\u4F9D\u8D56"), deps.map((d) => chip(String(d), "mute")).join(" ")]);
    if (cancelReq.has(taskIdOf(t2))) rows.push([t("\u53D6\u6D88"), chip(t("\u53D6\u6D88\u8BF7\u6C42\u5DF2\u53D1\u5E03"), "warn")]);
    el.innerHTML = '<div class="view-inner">' + head + '<div class="card"><div class="card-b">' + kv(rows) + (cancellable(t2) && !cancelReq.has(taskIdOf(t2)) ? '<div style="margin-top:10px"><button class="btn sm" data-task-cancel="' + esc(taskIdOf(t2)) + '">' + esc(t("\u53D6\u6D88")) + '</button><span class="small faint" style="margin-left:8px">' + esc(t("\u53D1\u5E03 TaskCancelRequested\u2014\u2014\u8C03\u5EA6\u5668\u5F52\u6863\u9762\u6D88\u8D39")) + "</span></div>" : "") + "</div></div></div>";
    el.querySelectorAll("[data-task-cancel]").forEach((btn) => {
      btn.addEventListener("click", () => requestCancel(el, ctx2, t2));
    });
  }
  var SEV_VIEW = { critical: { label: t("\u4E25\u91CD"), chip: "esc" }, high: { label: t("\u91CD\u8981"), chip: "warn" }, medium: { label: t("\u4E00\u822C"), chip: "mute" } };
  function sevView(sev) {
    return SEV_VIEW[String(sev || "")] || { label: String(sev || t("\u672A\u5206\u7EA7")), chip: "mute" };
  }
  function sevRank(sev) {
    const s = String(sev || "");
    return s === "critical" ? 0 : s === "high" ? 1 : s === "medium" ? 2 : 3;
  }
  function sortAttention(list2) {
    return (list2 || []).slice().sort((a, b) => sevRank(a && a.severity) - sevRank(b && b.severity));
  }
  function evidenceRoute(ref) {
    const r = String(ref || "").trim();
    if (!r) return "";
    if (r.startsWith("#/")) {
      const segs = r.slice(2).split("/").filter(Boolean);
      if (!segs.length) return "";
      if (segs[0] === "inbox") return "inbox";
      if (segs[0] === "approvals") return "approvals";
      if (segs[0] === "tasks") return "tasks";
      if (["s", "wf"].includes(segs[0]) && segs.length >= 2) return segs.map((x) => encodeURIComponent(x)).join("/");
      if (["event", "task", "session"].includes(segs[0]) && segs.length >= 2) return segs[0] + "/" + encodeURIComponent(segs.slice(1).join("/"));
      if (segs[0] === "obj" && segs.length >= 4) return segs.map((x) => encodeURIComponent(x)).join("/");
      return "";
    }
    return involvedRoute2(r);
  }
  function attentionCard(a) {
    if (!a) return "";
    const sev = sevView(a.severity);
    const evs = (a.evidence || []).map((ev) => {
      const ref = String(ev && ev.ref || "");
      const route2 = evidenceRoute(ref);
      return route2 ? chipGo(ref, route2, "mute", true, t("\u8BC1\u636E\u2014\u2014\u70B9\u51FB\u8FDB\u5165")) : chip(ref || ev && ev.kind || t("\u8BC1\u636E"), "mute");
    }).filter(Boolean);
    const action = a.action || {};
    const route = action.target ? String(action.target).replace(/^#\/?/, "") : "";
    return '<div class="att-card sev-' + esc(String(a.severity || "unknown")) + '"><div class="att-h">' + chip(sev.label, sev.chip) + (a.category ? chip(a.category, "mute") : "") + "<b>" + esc(String(a.title || t("\u672A\u547D\u540D\u5173\u6CE8\u9879"))) + "</b></div>" + (a.why ? '<div class="att-why small">' + esc(String(a.why)) + "</div>" : "") + '<div class="att-ev">' + (evs.length ? evs.join(" ") : t('<span class="small faint">\u8BC1\u636E\uFF1A\u2014</span>')) + "</div>" + (route ? '<div class="att-acts"><button class="btn sm" data-go="' + esc(route) + '">' + esc(String(action.label || t("\u53BB\u5904\u7406"))) + "</button></div>" : "") + "</div>";
  }
  function reportFooter(body) {
    const b = body || {};
    const dq = b.dataQuality || {};
    const missing = dq.missing;
    const missText = Array.isArray(missing) ? missing.map((m) => String(m)).join("\u3001") : String(missing || "");
    const gb = String(b.generatedBy || "");
    const gbLabel = gb === "llm" ? t("\u667A\u80FD\u5F52\u7EB3") : gb === "rules" ? t("\u57FA\u7840\u6C47\u603B") : gb || t("\u672A\u6CE8\u660E\u751F\u6210\u65B9\u5F0F");
    return t('<div class="report-foot small faint">\u6570\u636E\u53EF\u4FE1\u5EA6\uFF1A') + esc(gbLabel) + t(" \xB7 \u68C0\u67E5\u8303\u56F4 ") + esc(String(b.scope || "global")) + (b.asOf ? t(" \xB7 \u622A\u81F3 ") + esc(String(b.asOf)) : "") + (missText ? t(" \xB7 \u7F3A\u5E2D\u6570\u636E\uFF1A") + esc(missText) : "") + t(" \xB7 \u62A5\u544A\u4E3A\u6D3E\u751F\u89C6\u56FE\u53EF\u968F\u65F6\u91CD\u5EFA\uFF0C\u4E0D\u843D\u5BF9\u8C61\u8D26\u3002</div>");
  }
  function reportScreen(el, ctx2) {
    const st = ctx2.store;
    let att = st.attention();
    let degradedNote = "";
    if (!att.length) {
      const evs = st.events().filter((e) => e && (e.type === "ApprovalNeeded" || e.type === "Escalate"));
      att = evs.map((e) => ({
        severity: e.type === "Escalate" ? "high" : "medium",
        category: e.type,
        title: String(e.summary || e.type),
        evidence: e.involved ? [{ ref: String(e.involved) }] : []
      }));
      if (!evs.length) degradedNote = t('<div class="small faint" style="margin-bottom:10px">\u5FEB\u7167\u672A\u643A\u5E26 attention \u805A\u5408\uFF0C\u4E8B\u4EF6\u5F52\u6863\u4EA6\u65E0\u5BA1\u6279/\u5347\u7EA7\u884C\u2014\u2014\u672C\u8F6E\u62A5\u544A\u6309\u5728\u518C\u4E8B\u5B9E\u6536\u655B\u4E3A\u96F6\u547D\u4E2D\uFF08\u4E0D\u88C5\u5047\uFF09\u3002</div>');
      else degradedNote = t('<div class="small faint" style="margin-bottom:10px">\u5FEB\u7167\u672A\u643A\u5E26 attention \u805A\u5408\u2014\u2014\u672C\u5361\u7531\u4E8B\u4EF6\u5F52\u6863\u884C\uFF08ApprovalNeeded/Escalate\uFF09\u6295\u5F71\u751F\u6210\u3002</div>');
    }
    const sorted = sortAttention(att);
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u5173\u6CE8\u62A5\u544A"), t("severity \u6392\u5E8F \xB7 \u8BC1\u636E\u53EF\u70B9 \xB7 \u6570\u636E\u53EF\u4FE1\u5EA6\u663E\u5F0F"), chipGo(t("\u4EFB\u52A1\u4E2D\u5FC3"), "tasks", "mute", false)) + degradedNote + '<div data-loop-box="pending"></div>' + (sorted.length ? '<div class="att-list">' + sorted.map(attentionCard).join("") + "</div>" : emptyBox(t("\u672C\u7A97\u53E3\u65E0\u5173\u6CE8\u9879"), t("attention \u805A\u5408\u96F6\u547D\u4E2D\uFF08\u7A7A\u6001\u663E\u5F0F\u2014\u2014\u4E0D\u5192\u5145\u6709\u98CE\u9669\uFF09\u3002"))) + reportFooter(st.get("attention") && st.get("attention").dataQuality) + "</div>";
    hydrateLoop(el, ctx2);
  }
  function render6(el, ctx2) {
    if (ctx2.params && ctx2.params.id != null && ctx2.params.id !== "") taskDetailScreen(el, ctx2, String(ctx2.params.id));
    else tasksScreen(el, ctx2);
  }
  function renderReport(el, ctx2) {
    reportScreen(el, ctx2);
  }

  // node_modules/diff/libesm/diff/base.js
  var Diff = class {
    diff(oldStr, newStr, options2 = {}) {
      let callback;
      if (typeof options2 === "function") {
        callback = options2;
        options2 = {};
      } else if ("callback" in options2) {
        callback = options2.callback;
      }
      const oldString = this.castInput(oldStr, options2);
      const newString = this.castInput(newStr, options2);
      const oldTokens = this.removeEmpty(this.tokenize(oldString, options2));
      const newTokens = this.removeEmpty(this.tokenize(newString, options2));
      return this.diffWithOptionsObj(oldTokens, newTokens, options2, callback);
    }
    diffWithOptionsObj(oldTokens, newTokens, options2, callback) {
      var _a2;
      const done = (value) => {
        value = this.postProcess(value, options2);
        if (callback) {
          setTimeout(function() {
            callback(value);
          }, 0);
          return void 0;
        } else {
          return value;
        }
      };
      const newLen = newTokens.length, oldLen = oldTokens.length;
      let editLength = 1;
      let maxEditLength = newLen + oldLen;
      if (options2.maxEditLength != null) {
        maxEditLength = Math.min(maxEditLength, options2.maxEditLength);
      }
      const maxExecutionTime = (_a2 = options2.timeout) !== null && _a2 !== void 0 ? _a2 : Infinity;
      const abortAfterTimestamp = Date.now() + maxExecutionTime;
      const bestPath = [{ oldPos: -1, lastComponent: void 0 }];
      let newPos = this.extractCommon(bestPath[0], newTokens, oldTokens, 0, options2);
      if (bestPath[0].oldPos + 1 >= oldLen && newPos + 1 >= newLen) {
        return done(this.buildValues(bestPath[0].lastComponent, newTokens, oldTokens));
      }
      let minDiagonalToConsider = -Infinity, maxDiagonalToConsider = Infinity;
      const execEditLength = () => {
        for (let diagonalPath = Math.max(minDiagonalToConsider, -editLength); diagonalPath <= Math.min(maxDiagonalToConsider, editLength); diagonalPath += 2) {
          let basePath;
          const removePath = bestPath[diagonalPath - 1], addPath = bestPath[diagonalPath + 1];
          if (removePath) {
            bestPath[diagonalPath - 1] = void 0;
          }
          let canAdd = false;
          if (addPath) {
            const addPathNewPos = addPath.oldPos - diagonalPath;
            canAdd = addPath && 0 <= addPathNewPos && addPathNewPos < newLen;
          }
          const canRemove = removePath && removePath.oldPos + 1 < oldLen;
          if (!canAdd && !canRemove) {
            bestPath[diagonalPath] = void 0;
            continue;
          }
          if (!canRemove || canAdd && removePath.oldPos < addPath.oldPos) {
            basePath = this.addToPath(addPath, true, false, 0, options2);
          } else {
            basePath = this.addToPath(removePath, false, true, 1, options2);
          }
          newPos = this.extractCommon(basePath, newTokens, oldTokens, diagonalPath, options2);
          if (basePath.oldPos + 1 >= oldLen && newPos + 1 >= newLen) {
            return done(this.buildValues(basePath.lastComponent, newTokens, oldTokens)) || true;
          } else {
            bestPath[diagonalPath] = basePath;
            if (basePath.oldPos + 1 >= oldLen) {
              maxDiagonalToConsider = Math.min(maxDiagonalToConsider, diagonalPath - 1);
            }
            if (newPos + 1 >= newLen) {
              minDiagonalToConsider = Math.max(minDiagonalToConsider, diagonalPath + 1);
            }
          }
        }
        editLength++;
      };
      if (callback) {
        (function exec() {
          setTimeout(function() {
            if (editLength > maxEditLength || Date.now() > abortAfterTimestamp) {
              return callback(void 0);
            }
            if (!execEditLength()) {
              exec();
            }
          }, 0);
        })();
      } else {
        while (editLength <= maxEditLength && Date.now() <= abortAfterTimestamp) {
          const ret = execEditLength();
          if (ret) {
            return ret;
          }
        }
      }
    }
    addToPath(path, added, removed, oldPosInc, options2) {
      const last = path.lastComponent;
      if (last && !options2.oneChangePerToken && last.added === added && last.removed === removed) {
        return {
          oldPos: path.oldPos + oldPosInc,
          lastComponent: { count: last.count + 1, added, removed, previousComponent: last.previousComponent }
        };
      } else {
        return {
          oldPos: path.oldPos + oldPosInc,
          lastComponent: { count: 1, added, removed, previousComponent: last }
        };
      }
    }
    extractCommon(basePath, newTokens, oldTokens, diagonalPath, options2) {
      const newLen = newTokens.length, oldLen = oldTokens.length;
      let oldPos = basePath.oldPos, newPos = oldPos - diagonalPath, commonCount = 0;
      while (newPos + 1 < newLen && oldPos + 1 < oldLen && this.equals(oldTokens[oldPos + 1], newTokens[newPos + 1], options2)) {
        newPos++;
        oldPos++;
        commonCount++;
        if (options2.oneChangePerToken) {
          basePath.lastComponent = { count: 1, previousComponent: basePath.lastComponent, added: false, removed: false };
        }
      }
      if (commonCount && !options2.oneChangePerToken) {
        basePath.lastComponent = { count: commonCount, previousComponent: basePath.lastComponent, added: false, removed: false };
      }
      basePath.oldPos = oldPos;
      return newPos;
    }
    equals(left, right, options2) {
      if (options2.comparator) {
        return options2.comparator(left, right);
      } else {
        return left === right || !!options2.ignoreCase && left.toLowerCase() === right.toLowerCase();
      }
    }
    removeEmpty(array) {
      const ret = [];
      for (let i = 0; i < array.length; i++) {
        if (array[i]) {
          ret.push(array[i]);
        }
      }
      return ret;
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    castInput(value, options2) {
      return value;
    }
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    tokenize(value, options2) {
      return Array.from(value);
    }
    join(chars) {
      return chars.join("");
    }
    postProcess(changeObjects, options2) {
      return changeObjects;
    }
    get useLongestToken() {
      return false;
    }
    buildValues(lastComponent, newTokens, oldTokens) {
      const components = [];
      let nextComponent;
      while (lastComponent) {
        components.push(lastComponent);
        nextComponent = lastComponent.previousComponent;
        delete lastComponent.previousComponent;
        lastComponent = nextComponent;
      }
      components.reverse();
      const componentLen = components.length;
      let componentPos = 0, newPos = 0, oldPos = 0;
      for (; componentPos < componentLen; componentPos++) {
        const component = components[componentPos];
        if (!component.removed) {
          if (!component.added && this.useLongestToken) {
            let value = newTokens.slice(newPos, newPos + component.count);
            value = value.map(function(value2, i) {
              const oldValue = oldTokens[oldPos + i];
              return oldValue.length > value2.length ? oldValue : value2;
            });
            component.value = this.join(value);
          } else {
            component.value = this.join(newTokens.slice(newPos, newPos + component.count));
          }
          newPos += component.count;
          if (!component.added) {
            oldPos += component.count;
          }
        } else {
          component.value = this.join(oldTokens.slice(oldPos, oldPos + component.count));
          oldPos += component.count;
        }
      }
      return components;
    }
  };

  // node_modules/diff/libesm/diff/line.js
  var LineDiff = class extends Diff {
    constructor() {
      super(...arguments);
      this.tokenize = tokenize;
    }
    equals(left, right, options2) {
      if (options2.ignoreWhitespace) {
        if (!options2.newlineIsToken || !left.includes("\n")) {
          left = left.trim();
        }
        if (!options2.newlineIsToken || !right.includes("\n")) {
          right = right.trim();
        }
      } else if (options2.ignoreNewlineAtEof && !options2.newlineIsToken) {
        if (left.endsWith("\n")) {
          left = left.slice(0, -1);
        }
        if (right.endsWith("\n")) {
          right = right.slice(0, -1);
        }
      }
      return super.equals(left, right, options2);
    }
  };
  var lineDiff = new LineDiff();
  function diffLines(oldStr, newStr, options2) {
    return lineDiff.diff(oldStr, newStr, options2);
  }
  function tokenize(value, options2) {
    if (options2.stripTrailingCr) {
      value = value.replace(/\r\n/g, "\n");
    }
    const retLines = [], linesAndNewlines = value.split(/(\n|\r\n)/);
    if (!linesAndNewlines[linesAndNewlines.length - 1]) {
      linesAndNewlines.pop();
    }
    for (let i = 0; i < linesAndNewlines.length; i++) {
      const line = linesAndNewlines[i];
      if (i % 2 && !options2.newlineIsToken) {
        retLines[retLines.length - 1] += line;
      } else {
        retLines.push(line);
      }
    }
    return retLines;
  }

  // src/render/diff.js
  function diffLines2(oldStr, newStr) {
    const parts = diffLines(String(oldStr == null ? "" : oldStr), String(newStr == null ? "" : newStr));
    const rows = [];
    for (const p of parts) {
      const lines = String(p.value == null ? "" : p.value).split("\n");
      if (lines.length && lines[lines.length - 1] === "") lines.pop();
      const t2 = p.added ? "+" : p.removed ? "-" : " ";
      for (const s of lines) rows.push({ t: t2, s });
    }
    return compress(rows);
  }
  function compress(rows) {
    const keep = new Array(rows.length).fill(false);
    rows.forEach((r, i) => {
      if (r.t !== " ") {
        for (let j = Math.max(0, i - 2); j <= Math.min(rows.length - 1, i + 2); j++) keep[j] = true;
      }
    });
    const out = [];
    let last = -3;
    rows.forEach((r, i) => {
      if (!keep[i]) return;
      if (i - last > 1 && r.t === " ") out.push({ t: "gap", s: "\u22EF" });
      out.push(r);
      last = i;
    });
    if (!out.length) out.push({ t: "ctx", s: t("\uFF08\u65E0\u5DEE\u5F02\uFF09") });
    return out;
  }
  function renderDiff(lines, maxLines = 200) {
    const cap = Number(maxLines) || 200;
    const html3 = lines.slice(0, cap).map((l) => {
      if (l.t === "gap") return '<div class="diff-line ctx">   \u22EF</div>';
      const sign = l.t === "+" ? "+" : l.t === "-" ? "-" : " ";
      const cls = l.t === "+" ? "add" : l.t === "-" ? "del" : "ctx";
      return '<div class="diff-line ' + cls + '"><span class="d-sign">' + sign + "</span>" + esc(trunc(l.s, 500)) + "</div>";
    }).join("");
    const tail = lines.length > cap ? t('<div class="diff-line ctx">\u2026 \u5171 ') + lines.length + t(" \u884C\u5DEE\u5F02</div>") : "";
    return '<div class="diff">' + html3 + tail + "</div>";
  }
  function diffStats(lines) {
    let add = 0;
    let del = 0;
    (lines || []).forEach((l) => {
      if (l.t === "+") add += 1;
      else if (l.t === "-") del += 1;
    });
    return { add, del };
  }
  function hunksOf(lines) {
    let n = 0;
    let inHunk = false;
    (lines || []).forEach((l) => {
      if (l.t === "+" || l.t === "-") {
        if (!inHunk) {
          n += 1;
          inHunk = true;
        }
      } else {
        inHunk = false;
      }
    });
    return n;
  }

  // src/screens/editor.js
  function splitFrontmatter(content) {
    const s = String(content || "");
    if (!s.startsWith("---")) return { fm: "", body: s };
    const end = s.indexOf("\n---", 3);
    if (end < 0) return { fm: "", body: s };
    return { fm: s.slice(0, end + 4), body: s.slice(end + 4).replace(/^\n/, "") };
  }
  function aiFaces() {
    return typeof window !== "undefined" && window.__CC && window.__CC.chatFaces || null;
  }
  var ED = /* @__PURE__ */ new Map();
  var saveTimers = /* @__PURE__ */ new Map();
  var notice = "";
  var screenCtx = null;
  function keyOf(kind, scope, name) {
    return kind + "|" + (scope || "global") + "|" + (name || "");
  }
  function hasDiff(ed) {
    return !!ed && ed.loaded && ed.working !== ed.base;
  }
  function resolveRef(ctx2) {
    const p = ctx2.params || {};
    const objs = ctx2.store.objects();
    if (p.kind && p.name) {
      const hit = objs.find((o) => o.kind === p.kind && o.name === p.name && (!p.scope || p.scope === "global" || o.scope === p.scope));
      if (hit) return { kind: hit.kind, scope: hit.scope, name: hit.name, rv: hit.rv };
      return { kind: p.kind, scope: p.scope || "global", name: p.name, rv: 0 };
    }
    if (p.name) {
      const cands = objs.filter((o) => o.name === p.name);
      if (cands.length === 1) return { kind: cands[0].kind, scope: cands[0].scope, name: cands[0].name, rv: cands[0].rv };
      return null;
    }
    return null;
  }
  async function hydrate(ctx2, key) {
    const { kind, scope, name, rv } = key.ref;
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const faceP = ctx2.api.object(kind, scope, name);
    const row = ctx2.store.drafts().find((d) => d && d.kind === kind && (d.scope || "global") === scope && d.name === name);
    const draftP = row ? ctx2.api.draft(kind, name, scope) : Promise.resolve({ ok: false, status: 404, error: t("\u65E0\u7A3F") });
    const [faceR, draftR] = await Promise.all([faceP, draftP]);
    if ((typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    const prev = ED.get(key.key);
    let base = "";
    let faceRV = 0;
    let faceAbsent = false;
    let faceFail = "";
    if (faceR.ok) {
      const data = faceR.data && typeof faceR.data === "object" ? faceR.data : {};
      base = data.content != null ? String(data.content) : typeof faceR.data === "string" ? faceR.data : JSON.stringify(faceR.data, null, 2);
      faceRV = Number(data.resourceVersion) || rv || 0;
    } else if (faceR.status === 404 && row) {
      faceAbsent = true;
    } else {
      faceFail = String(faceR.error || faceR.status);
    }
    const working = draftR && draftR.ok && draftR.data && draftR.data.content != null ? String(draftR.data.content) : base;
    ED.set(key.key, {
      ref: key.ref,
      working,
      base,
      faceAbsent,
      faceRV,
      faceFail,
      baseRV: draftR && draftR.ok && draftR.data ? Number(draftR.data.baseResourceVersion) || faceRV : faceRV,
      draftFM: draftR && draftR.ok && draftR.data ? String(draftR.data.fieldManager || "") : "",
      view: prev && prev.view || "render",
      tab: prev && prev.tab || "draft",
      loaded: true
    });
    paintBody(key.key);
    paintBadges(key.key);
    paintRail(key.key, ctx2);
    paintMeta(key.key);
  }
  function bindRetry(keyStr) {
    document.querySelectorAll("[data-ed-retry]").forEach((b) => {
      if (b.getAttribute("data-ed-retry-bound")) return;
      b.setAttribute("data-ed-retry-bound", "1");
      b.addEventListener("click", () => {
        const cur = ED.get(keyStr);
        if (!cur || !screenCtx) return;
        ED.delete(keyStr);
        const body = document.querySelector("[data-ed-body]");
        if (body) body.innerHTML = '<div class="small faint">' + esc(t("\u6CE8\u6C34\u4E2D\u2026")) + "</div>";
        paintMeta(keyStr);
        hydrate(screenCtx, { ref: cur.ref, key: keyStr });
      });
    });
  }
  function paintMeta(keyStr) {
    const slot = document.querySelector("[data-ed-rv]");
    const ed = ED.get(keyStr);
    if (!slot) return;
    if (!ed) slot.innerHTML = esc(t("\u6CE8\u6C34\u4E2D\u2026"));
    else if (ed.faceFail) {
      slot.innerHTML = '<span class="chip esc tiny">' + esc(t("\u6B63\u6587\u62C9\u53D6\u5931\u8D25\uFF08") + ed.faceFail + t("\uFF09")) + '</span> <button class="btn sm" data-ed-retry>' + esc(t("\u91CD\u8BD5")) + "</button>";
      bindRetry(keyStr);
    } else if (ed.faceAbsent) slot.innerHTML = esc(t("\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09"));
    else slot.innerHTML = "r" + ed.faceRV + esc(t("\uFF08\u4E50\u89C2\u9501\u57FA\u51C6\uFF09"));
  }
  function paintBody(key) {
    const el = document.querySelector("[data-ed-body]");
    const ed = ED.get(key);
    if (!el || !ed) return;
    if (ed.faceFail) {
      el.innerHTML = '<div class="empty"><div class="e-title">' + esc(t("\u6B63\u6587\u62C9\u53D6\u5931\u8D25\uFF08") + ed.faceFail + t("\uFF09")) + '</div><div class="small faint">' + esc(t("face \u73B0\u8BFB\u672A\u6210\u529F\u2014\u2014\u4E50\u89C2\u9501\u57FA\u51C6\u672A\u77E5\uFF0C\u7F16\u8F91/Apply \u5DF2\u505C\u7528\uFF08\u96F6\u88C5\u5047\uFF09\uFF1B\u70B9\u91CD\u8BD5\u91CD\u8DD1\u6CE8\u6C34\u3002")) + '</div><div class="e-act"><button class="btn sm primary" data-ed-retry>' + esc(t("\u91CD\u8BD5")) + "</button></div></div>";
      bindRetry(key);
      return;
    }
    el.innerHTML = ed.view === "render" ? isMarkdownDoc(ed.ref.kind, ed.ref.name) ? (function() {
      const { fm, body } = splitFrontmatter(ed.working);
      const fmHtml = fm ? '<details class="scheme-fold"><summary class="ex-sum"><span class="ex-chev">\u25B8</span>' + esc(t("scheme\uFF08frontmatter \u58F0\u660E\u2014\u2014expand \u67E5\u770B\u5B8C\u6574\uFF09")) + '</summary><pre class="mono small scheme-pre">' + esc(fm) + "</pre></details>" : "";
      return '<div data-md-render="1">' + fmHtml + renderMarkdown(body) + "</div>" + rawDisclosure(t("\u539F\u6587"), ed.working);
    })() : '<div data-md-cards="1">' + (yamlToCards(ed.working) || t('<div class="ob-empty">\uFF08\u7A7A\u58F0\u660E\uFF09</div>')) + "</div>" + rawDisclosure("YAML", ed.working) : "";
    const ta = document.querySelector("[data-ed-textarea]");
    if (ta) {
      ta.hidden = ed.view !== "raw";
      if (!ta.hidden && ta.value !== ed.working && document.activeElement !== ta) ta.value = ed.working;
    }
    const renderWrap = document.querySelector('[data-ed-view="render"]');
    if (renderWrap) renderWrap.hidden = ed.view !== "render";
  }
  function paintBadges(key) {
    const ed = ED.get(key);
    if (!ed) return;
    const has = hasDiff(ed);
    const hunks = has ? hunksOf(diffLines2(ed.base, ed.working)) : 0;
    const chipEl = document.querySelector("[data-ed-draft-chip]");
    if (chipEl) {
      chipEl.className = has ? "chip warn tiny" : "chip ok tiny";
      chipEl.textContent = has ? t("\u25CF \u672A Apply\uFF08\u8349\u7A3F ") + hunks + t(" \u5904\u53D8\u66F4\uFF09") : t("\u65E0\u8349\u7A3F");
    }
    const dock = document.querySelector("[data-ed-apply]");
    if (dock) {
      dock.disabled = !has;
      dock.textContent = has ? "Apply\uFF08" + hunks + t(" \u5904\u53D8\u66F4\uFF09") : t("Apply\uFF08\u65E0\u53D8\u66F4\uFF09");
    }
  }
  function paintRail(key, ctx2) {
    const box = document.querySelector("[data-ed-rail]");
    const ed = ED.get(key);
    if (!box || !ed) return;
    if (ed.tab === "ai") {
      const f = aiFaces();
      const canAi = !!(f && typeof f.docFix === "function");
      box.innerHTML = t('<div class="card-h"><span class="h-title">AI \u4FEE\u6587\u6863</span></div><div class="card-b">') + '<div class="small faint">' + esc(t("\u6D41\u7A0B\uFF1A\u5199\u4FEE\u6539\u610F\u56FE \u2192 AI \u51FA\u5168\u6587\u5EFA\u8BAE \u2192 \u5199\u5165\u8349\u7A3F \u2192 \u53F3\u4E0B Apply\uFF08\u4E50\u89C2\u9501+\u9884\u68C0\u7167\u8D70\u2014\u2014AI \u65E0\u7279\u6743\uFF09\u3002")) + '</div><textarea class="inp" data-ai-prompt rows="3" style="width:100%;margin-top:8px" placeholder="' + esc(t("\u4FEE\u6539\u610F\u56FE\uFF0C\u5982\uFF1A\u8865\u5145 X \u8282\uFF1B\u4FEE\u6B63 Y \u5904\u53E3\u5F84\u2026")) + '"></textarea><button class="btn primary sm" data-ai-ask style="margin-top:8px"' + (canAi ? "" : " disabled") + ">" + esc(t("\u8BF7\u6C42 AI \u5EFA\u8BAE")) + '</button><div data-ai-out style="margin-top:10px">' + (canAi ? '<div class="small faint">' + esc(t("\u5EFA\u8BAE\u5C06\u73B0\u8BFB\u4E8E\u6B64\u2014\u2014\u5199\u5165\u8349\u7A3F\u540E\u8D70\u65E2\u6709 Apply \u901A\u8DEF\u3002")) + "</div>" : '<div class="small faint">' + esc(t("AI \u901A\u9053\u672A\u6CE8\u5165\uFF08__CC.chatFaces.docFix \u7F3A\u5E2D\uFF09\u2014\u2014\u53EF\u5728\u300C\u539F\u6587\u300D\u89C6\u56FE\u624B\u7F16\uFF0C\u6216\u7ECF\u4F1A\u8BDD\u5C4F\u5BF9\u8BDD\u540E\u4EBA\u5DE5\u8D34\u56DE\u3002"))) + "</div></div>";
      const askBtn = box.querySelector("[data-ai-ask]");
      if (askBtn) askBtn.addEventListener("click", async () => {
        const promptTa = box.querySelector("[data-ai-prompt]");
        const out = box.querySelector("[data-ai-out]");
        if (!promptTa || !out) return;
        const intent = promptTa.value.trim();
        if (!intent) return;
        askBtn.disabled = true;
        out.innerHTML = '<div class="small faint">' + esc(t("AI \u73B0\u8BFB\u4E2D\u2026")) + "</div>";
        try {
          const sug = await f.docFix(intent, { kind: ed.ref.kind, scope: ed.ref.scope, name: ed.ref.name, base: ed.working });
          const text2 = String(sug == null ? "" : sug);
          if (!text2.trim()) throw new Error(t("\u7A7A\u5EFA\u8BAE"));
          out.innerHTML = '<div class="small faint">' + esc(t("AI \u5EFA\u8BAE\u5DF2\u5230\uFF08")) + text2.length + ' B\uFF09</div><pre class="mono small" style="max-height:220px;overflow:auto;border:1px solid var(--bd,#2a2f3a);padding:8px;border-radius:8px">' + esc(trunc(text2, 4e3)) + '</pre><button class="btn primary sm" data-ai-apply style="margin-top:8px">' + esc(t("\u5199\u5165\u8349\u7A3F")) + "</button>";
          out.querySelector("[data-ai-apply]").addEventListener("click", () => {
            ed.working = text2;
            paintBody(ed.ref.kind + "/" + (ed.ref.scope || "global") + "/" + ed.ref.name);
            paintBadges(ed.ref.kind + "/" + (ed.ref.scope || "global") + "/" + ed.ref.name);
            out.innerHTML = '<div class="small faint">' + esc(t("\u5DF2\u5199\u5165\u8349\u7A3F\u2014\u2014\u53F3\u4E0B Apply \u843D\u5E93\uFF08\u4E50\u89C2\u9501\u7167\u8D70\uFF09\u3002")) + "</div>";
          });
        } catch (e) {
          out.innerHTML = '<div class="small faint">' + esc(t("AI \u5EFA\u8BAE\u5931\u8D25\uFF08")) + esc(String(e && e.message || e)) + "\uFF09</div>";
        } finally {
          askBtn.disabled = false;
        }
      });
      return;
    }
    if (ed.tab === "history") {
      const rows = ctx2.store.events().filter((e) => e && String(e.involved || "").includes(ed.ref.name));
      box.innerHTML = t('<div class="card-h"><span class="h-title">\u5386\u53F2</span></div><div class="card-b">') + (rows.length ? rows.slice(0, 30).map((e) => '<div class="mini-row"><div class="mr-main"><div>' + esc(String(e.type || "")) + '</div><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 50)) + '</div></div><span class="w-ev-time mono">' + esc(String((e.ts || e.at || e.time || "").slice(0, 16))) + "</span></div>").join("") : t('<div class="small faint">\u8BE5\u5BF9\u8C61\u6682\u65E0\u5F52\u6863\u4E8B\u4EF6\u884C\uFF08DocApplied/DocReview \u7B49\u2014\u2014\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>')) + "</div>";
      return;
    }
    if (!hasDiff(ed)) {
      box.innerHTML = t('<div class="card-h"><span class="h-title">\u8349\u7A3F</span></div><div class="card-b">') + t('<div class="small faint">\u65E0\u8349\u7A3F\u2014\u2014\u7F16\u8F91\u300C\u539F\u6587\u300D\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08\u5199\u7A7F /v1/drafts\uFF0C\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\u3002</div>') + (ed.draftFM ? t('<div class="small faint" style="margin-top:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A') + esc(ed.draftFM) + "</div>" : "") + "</div>";
      return;
    }
    const lines = diffLines2(ed.base, ed.working);
    const stats = diffStats(lines);
    box.innerHTML = t('<div class="card-h"><span class="h-title">\u8349\u7A3F</span><span class="h-right">') + chip("+" + stats.add, "ok") + " " + chip("-" + stats.del, "esc") + '</span></div><div class="card-b">' + (ed.draftFM ? t('<div class="small faint" style="margin-bottom:6px">\u7A3F\u6765\u6E90 fieldManager\uFF1A') + esc(ed.draftFM) + "</div>" : "") + renderDiff(lines, 40) + t('<div class="small faint" style="margin-top:6px">\u57FA\u51C6=r') + ed.baseRV + t("\uFF08\u4E50\u89C2\u9501\u671F\u671B\u503C\u2014\u2014409 \u4E0D\u9759\u9ED8\u91CD\u57FA\uFF09\xB7 \u7A3F ") + ed.working.length + t("B vs \u843D\u5E93\u503C ") + ed.base.length + "B</div></div>";
  }
  function scheduleSave(ctx2, key) {
    const prev = saveTimers.get(key.key);
    if (prev) clearTimeout(prev);
    saveTimers.set(key.key, setTimeout(async () => {
      const ed = ED.get(key.key);
      if (!ed || !hasDiff(ed)) return;
      const r = await ctx2.api.saveDraft({
        kind: ed.ref.kind,
        scope: ed.ref.scope,
        name: ed.ref.name,
        content: ed.working,
        baseResourceVersion: ed.baseRV,
        fieldManager: "console:web"
      });
      if (!r.ok) {
        notice = t("\u8349\u7A3F\u4FDD\u5B58\u5931\u8D25\uFF08") + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u5DE5\u4F5C\u526F\u672C\u4ECD\u5728\u5185\u5B58\uFF0C\u9000\u51FA\u524D\u4E0D\u4E22");
        paintNotice();
      }
    }, 800));
  }
  function paintNotice() {
    const slot = document.querySelector("[data-ed-notice]");
    if (slot) slot.innerHTML = notice ? '<div class="small faint">' + notice + "</div>" : "";
  }
  function editorScreen(el, ctx2, key) {
    const ed = ED.get(key.key);
    const view = ed ? ed.view : "render";
    el.innerHTML = '<div class="view-inner"><div class="doc-layout"><div class="doc-main"><div class="doc-headbar"><span class="doc-path mono">' + esc(key.ref.scope + "/" + key.ref.name) + "</span>" + kindBadge(key.ref.kind) + t('<span class="chip ok tiny" data-ed-draft-chip>\u65E0\u8349\u7A3F</span>') + '<div class="md-switch" style="margin-left:auto"><span class="chip click small' + (view === "render" ? " acc" : " mute") + t('" data-ed-view-btn="render">\u6B63\u6587</span>') + '<span class="chip click small' + (view === "raw" ? " acc" : " mute") + t('" data-ed-view-btn="raw">\u539F\u6587</span>') + '</div></div><div class="doc-meta-row">' + chip("scope", "mute") + chipGo(key.ref.scope, "s/" + encodeURIComponent(key.ref.scope), "mute") + '<span class="mono small" data-ed-rv>' + (ed ? ed.faceAbsent ? t("\u6587\u6863\u7F3A\u5E2D\uFF08\u5728\u518C\u7A3F\u57FA\u5EA7\u2014\u2014Apply=\u521B\u5EFA\u901A\u8DEF\uFF09") : "r" + ed.faceRV + t("\uFF08\u4E50\u89C2\u9501\u57FA\u51C6\uFF09") : t("\u6CE8\u6C34\u4E2D\u2026")) + '</span></div><div data-ed-notice></div><div class="doc-render-wrap" data-ed-view="render" data-ed-body' + (view === "render" ? "" : " hidden") + "></div>" + t('<textarea class="doc-editor" data-ed-textarea spellcheck="false" placeholder="\u539F\u6587\u7F16\u8F91\u2014\u2014\u53D8\u66F4\u5373\u7D2F\u79EF\u8349\u7A3F"') + (view === "raw" ? "" : " hidden") + "></textarea>" + t('<div class="small faint">\u9ED8\u8BA4\u5C55\u793A\u6E32\u67D3\u540E\u7684\u6210\u54C1\uFF1B\u70B9\u300C\u539F\u6587\u300D\u67E5\u770B\u4E0E\u7F16\u8F91\u6E90\u6587\u4EF6\u3002\u7F16\u8F91\u5373\u7D2F\u79EF\u8349\u7A3F\uFF08POST /v1/drafts\u2014\u2014\u4E0D\u52A8\u6B63\u5F0F\u6587\u6863\u3001\u4E0D\u53D1\u4E8B\u4EF6\uFF09\uFF1BApply \u552F\u4E00\u5165\u53E3=\u53F3\u680F\u5438\u5E95\u6309\u94AE\u3002</div>') + '</div><div class="doc-collab"><div class="collab-tabs"><div class="collab-tab' + ((ed && ed.tab) === "draft" ? " active" : "") + t('" data-ed-tab="draft">\u8349\u7A3F</div>') + '<div class="collab-tab' + ((ed && ed.tab) === "history" ? " active" : "") + t('" data-ed-tab="history">\u5386\u53F2</div>') + '<div class="collab-tab' + ((ed && ed.tab) === "ai" ? " active" : "") + t('" data-ed-tab="ai">AI \u4FEE\u6587\u6863</div>') + '</div><div class="collab-body" data-ed-rail></div>' + t('<div class="apply-dock"><button class="btn primary block" data-ed-apply disabled>Apply\uFF08\u65E0\u53D8\u66F4\uFF09</button></div>') + "</div></div></div>";
    paintNotice();
    paintMeta(key.key);
    if (ed && ed.loaded) {
      paintBody(key.key);
      paintBadges(key.key);
      paintRail(key.key, ctx2);
    } else hydrate(ctx2, key);
    el.querySelectorAll("[data-ed-view-btn]").forEach((b) => {
      b.addEventListener("click", () => {
        const cur = ED.get(key.key);
        if (!cur) return;
        cur.view = b.dataset.edViewBtn;
        el.querySelectorAll("[data-ed-view-btn]").forEach((x) => {
          x.className = "chip click small" + (x.dataset.edViewBtn === cur.view ? " acc" : " mute");
        });
        paintBody(key.key);
      });
    });
    el.querySelectorAll("[data-ed-tab]").forEach((t2) => {
      t2.addEventListener("click", () => {
        const cur = ED.get(key.key);
        if (!cur) return;
        cur.tab = t2.dataset.edTab;
        el.querySelectorAll("[data-ed-tab]").forEach((x) => x.classList.toggle("active", x === t2));
        paintRail(key.key, ctx2);
      });
    });
    const ta = el.querySelector("[data-ed-textarea]");
    if (ta) {
      if (ed && ed.loaded && document.activeElement !== ta) ta.value = ed.working;
      ta.addEventListener("input", () => {
        const cur = ED.get(key.key);
        if (!cur) return;
        cur.working = ta.value;
        paintBadges(key.key);
        paintRail(key.key, ctx2);
        scheduleSave(ctx2, key);
      });
    }
    const dock = el.querySelector("[data-ed-apply]");
    if (dock) {
      dock.addEventListener("click", async () => {
        const cur = ED.get(key.key);
        if (!cur || !hasDiff(cur)) return;
        const stats = diffStats(diffLines2(cur.base, cur.working));
        if (!window.confirm(t("Apply \u843D\u5E93 ") + cur.ref.kind + "/" + cur.ref.name + "\uFF08+" + stats.add + "/-" + stats.del + t("\uFF09\uFF1F\u4E50\u89C2\u9501\u57FA\u51C6 r") + cur.baseRV + "\u3002")) return;
        dock.disabled = true;
        const r = await ctx2.api.apply({
          kind: cur.ref.kind,
          scope: cur.ref.scope,
          name: cur.ref.name,
          content: cur.working,
          baseResourceVersion: cur.baseRV,
          fieldManager: "console:web",
          reasonRef: "console:web-editor"
        });
        dock.disabled = false;
        if (!r.ok) {
          notice = t("Apply \u5931\u8D25\uFF08") + esc(String(r.error || r.status)) + "\uFF09" + (r.status === 409 ? t("\u2014\u2014\u4E50\u89C2\u9501\u51B2\u7A81\uFF1A\u7A3F\u5DF2\u4FDD\u7559\uFF0C\u8BF7\u5BF9\u9F50\u6700\u65B0 rv \u540E\u91CD\u8BD5") : "");
          paintNotice();
          return;
        }
        notice = t("Apply \u6210\u529F\u2014\u2014\u6587\u6863\u5DF2\u843D\u5E93\uFF08rv=") + esc(String(r.data && r.data.rv || "?")) + "\uFF09\u3002";
        cur.base = cur.working;
        paintBadges(key.key);
        paintRail(key.key, ctx2);
        paintNotice();
        ctx2.api.snapshot().then((s) => {
          if (s.ok) ctx2.store.init(s.data);
        });
      });
    }
  }
  function render7(el, ctx2) {
    screenCtx = ctx2;
    const ref = resolveRef(ctx2);
    if (!ref) {
      el.innerHTML = '<div class="view-inner">' + screenHead(t("\u6587\u6863\u7F16\u8F91\u5668"), "") + emptyBox(t("\u5BF9\u8C61\u89E3\u6790\u6B67\u4E49\u6216\u7F3A\u5E2D"), t("\u540C\u540D\u5BF9\u8C61\u8DE8 kind/scope \u4E0D\u552F\u4E00\u2014\u2014\u8BF7\u4ECE\u7CFB\u7EDF\u6811\u8FDB\u5165\u3002"), chipGo(t("\u6253\u5F00\u7CFB\u7EDF\u6811"), "tree", "primary", false)) + "</div>";
      return;
    }
    const key = { ref, key: keyOf(ref.kind, ref.scope, ref.name) };
    editorScreen(el, ctx2, key);
  }

  // src/screens/approvals.js
  var ignored = /* @__PURE__ */ new Set();
  function pendingItems(ctx2) {
    const drafts = ctx2.store.drafts().filter((d) => d && d.kind && d.name).map((d) => ({ kind: "draft", key: "draft|" + d.kind + "|" + (d.scope || "global") + "|" + d.name, draft: d }));
    const evs = ctx2.store.events().filter((e) => e && e.type === "ApprovalNeeded" && !ignored.has("ev|" + String(e.id ?? ""))).map((e) => ({ kind: "event", key: "ev|" + String(e.id ?? ""), ev: e }));
    return drafts.filter((x) => !ignored.has(x.key)).concat(evs);
  }
  function apKeyOf(d) {
    return (d.kind + "/" + (d.scope || "global") + "/" + d.name).replace(/[^A-Za-z0-9]/g, "_");
  }
  function wfAnchorOf(ev) {
    const first = String(ev && ev.involved || "").split(",")[0].trim();
    return /^Workflow\/[^/\s]+\/[^/\s]+$/.test(first) ? first : "";
  }
  function itemHtml(it) {
    if (it.kind === "event") {
      const e = it.ev;
      const involved = String(e.involved || "");
      const anchor = wfAnchorOf(e);
      const route = involvedRoute2(involved.split(",")[0] || "");
      const gateBtns = anchor ? '<button class="btn sm primary" data-ap-grant="' + esc(it.key) + '" data-ap-ev="' + esc(String(e.id == null ? "" : e.id)) + '" data-ap-anchor="' + esc(anchor) + '">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn sm" data-ap-reject="' + esc(it.key) + '" data-ap-ev="' + esc(String(e.id == null ? "" : e.id)) + '" data-ap-anchor="' + esc(anchor) + '">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button> " : '<button class="btn sm" disabled title="' + esc(t("\u4E8B\u4EF6\u672A\u643A\u5E26 Workflow \u951A")) + '">' + esc(t("\u6279\u51C6\u7EE7\u7EED")) + '</button> <button class="btn sm" disabled title="' + esc(t("\u4E8B\u4EF6\u672A\u643A\u5E26 Workflow \u951A")) + '">' + esc(t("\u9A73\u56DE\u7EC8\u6B62")) + "</button> ";
      return '<div class="card ap-item"><div class="card-h"><span class="h-title">' + esc(String(e.type)) + '</span><span class="h-right">' + chip(String(e.level || "warn"), "warn") + '</span></div><div class="card-b"><div class="small">' + esc(trunc(String(e.summary || e.involved || ""), 160)) + '</div><div style="margin-top:8px">' + (route ? chipGo(involved.split(",")[0], route, "mute", true, t("\u53BB\u5BF9\u8C61\u5DE5\u4F5C\u53F0")) : chip(involved || "\u2014", "mute")) + '</div><div class="dh-actions" style="margin-top:10px">' + gateBtns + '<button class="btn sm" data-ap-ignore="' + esc(it.key) + '">' + esc(t("\u5FFD\u7565")) + "</button></div></div></div>";
    }
    const d = it.draft;
    const scope = d.scope || "global";
    return '<div class="card ap-item" data-ap-item="' + esc(apKeyOf(d)) + '"><div class="card-h"><span class="h-title"><span class="lk mono" data-go="' + esc(docRoute(d.kind, scope, d.name)) + '">' + esc(scope + "/" + d.name) + '</span></span><span class="h-right">' + kindBadge(d.kind) + " " + chip(t("\u7A3F: ") + (d.fieldManager || "\u2014"), "mute") + '</span></div><div class="card-b"><div data-apdiff="' + esc(apKeyOf(d)) + t('"><div class="small faint">diff \u73B0\u8BFB\u4E2D\u2026</div></div>') + '<div class="dh-actions" style="margin-top:10px"><button class="btn sm primary" data-ap-approve="' + esc(apKeyOf(d)) + t('">\u6279\u51C6\uFF08= Apply\uFF09</button> ') + '<button class="btn sm" data-ap-ignore="' + esc(it.key) + t('">\u5FFD\u7565</button> ') + '<button class="btn sm" data-go="' + esc(docRoute(d.kind, scope, d.name)) + t('">\u7F16\u8F91\u5668</button>') + "</div></div></div>";
  }
  async function hydrateDiff(ctx2, d) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const [draftR, faceR] = await Promise.all([
      ctx2.api.draft(d.kind, d.name, d.scope || "global"),
      ctx2.api.object(d.kind, d.scope || "global", d.name)
    ]);
    if ((typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    const box = document.querySelector('[data-apdiff="' + apKeyOf(d) + '"]');
    if (!box) return;
    if (!draftR.ok) {
      box.innerHTML = t('<div class="small faint">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(draftR.error || draftR.status)) + t("\uFF09\u2014\u2014\u91CD\u8FDB\u5373\u6062\u590D\u3002</div>");
      return;
    }
    const dr = draftR.data && typeof draftR.data === "object" ? draftR.data : {};
    const working = String(dr.content || "");
    let base = t("\uFF08\u5BF9\u8C61\u4E0D\u5728\u518C\u2014\u2014apply \u5C06\u521B\u5EFA\u901A\u8DEF\uFF09");
    let baseNote = t("\u57FA\u51C6=r?");
    if (faceR.ok) {
      const data = faceR.data && typeof faceR.data === "object" ? faceR.data : {};
      base = data.content != null ? String(data.content) : "";
      baseNote = t("\u57FA\u51C6=r") + (dr.baseResourceVersion != null ? dr.baseResourceVersion : data.resourceVersion != null ? data.resourceVersion : "?");
    } else if (faceR.status !== 404) {
      base = t("\uFF08\u9762\u73B0\u503C\u6682\u4E0D\u53EF\u8BFB\uFF1A") + String(faceR.error || faceR.status) + "\uFF09";
    }
    const stats = diffStats(diffLines2(base, working));
    box.innerHTML = renderDiff(diffLines2(base, working), 26) + '<div class="small faint" style="margin-top:4px">' + esc(baseNote) + t(" \xB7 \u7A3F ") + working.length + t("B vs \u843D\u5E93\u503C ") + base.length + t("B\uFF08\u73B0\u8BFB\uFF0C\u4E0D\u7F13\u5B58\uFF09 \xB7 ") + '<span class="chip ok tiny">+' + stats.add + '</span> <span class="chip esc tiny">-' + stats.del + "</span></div>";
  }
  async function approve(ctx2, d, btn) {
    btn.disabled = true;
    btn.textContent = t("\u9884\u68C0\u4E2D\u2026");
    const dry = await ctx2.api.apply({
      kind: d.kind,
      scope: d.scope || "global",
      name: d.name,
      fromDraft: true,
      dryRun: true,
      fieldManager: "console:approve",
      reasonRef: "console:approve-dry"
    });
    btn.disabled = false;
    btn.textContent = t("\u6279\u51C6\uFF08= Apply\uFF09");
    if (!dry.ok) {
      if (typeof window !== "undefined") window.alert(t("dry-run \u9884\u68C0\u672A\u8FC7\uFF08") + String(dry.error || dry.status) + t("\uFF09\u2014\u2014\u96F6\u526F\u4F5C\u7528\uFF0C\u7A3F\u5DF2\u4FDD\u7559\u3002"));
      return;
    }
    const cur = dry.data && dry.data.rv != null ? Number(dry.data.rv) : null;
    const expect = d.baseResourceVersion != null ? Number(d.baseResourceVersion) : null;
    const lockNote = cur != null && expect != null ? cur === expect ? t("\u4E50\u89C2\u9501\u5BF9\u9F50 r") + cur : t("\u4E50\u89C2\u9501\u6F02\u79FB\uFF08\u73B0 r") + cur + t(" vs \u7A3F\u57FA r") + expect + t("\u2014\u2014\u843D\u5E93\u5C06\u88AB\u62D2\uFF09") : "";
    if (!window.confirm(t("\u786E\u8BA4\u6279\u51C6\uFF08Apply\uFF09") + d.kind + "/" + d.name + t("\uFF1Fdry-run \u5DF2\u8FC7\uFF08rv=") + String(cur) + "\uFF09" + (lockNote ? "\uFF1B" + lockNote : "") + "\u3002")) return;
    btn.disabled = true;
    btn.textContent = t("\u843D\u5E93\u4E2D\u2026");
    const r = await ctx2.api.apply({
      kind: d.kind,
      scope: d.scope || "global",
      name: d.name,
      fromDraft: true,
      fieldManager: "console:approve",
      reasonRef: "console:approve"
    });
    btn.disabled = false;
    btn.textContent = t("\u6279\u51C6\uFF08= Apply\uFF09");
    if (!r.ok) {
      window.alert(t("Apply \u5931\u8D25\uFF08") + String(r.error || r.status) + t("\uFF09\u2014\u2014\u7A3F\u4FDD\u7559\u8BED\u4E49\uFF0C\u672A\u843D\u5E93\u3002"));
      return;
    }
    ignored.add("draft|" + d.kind + "|" + (d.scope || "global") + "|" + d.name);
    const cardEl = document.querySelector('[data-ap-item="' + apKeyOf(d) + '"]');
    if (cardEl) cardEl.remove();
    ctx2.api.snapshot().then((s) => {
      if (s.ok) ctx2.store.init(s.data);
    });
  }
  async function pubGateDecision(ctx2, btn, grant) {
    const anchor = btn.getAttribute("data-ap-anchor") || "";
    const key = btn.getAttribute(grant ? "data-ap-grant" : "data-ap-reject") || "";
    const evId = btn.getAttribute("data-ap-ev") || "";
    if (!anchor || !key) return;
    const confirmed = grant ? window.confirm(t("\u786E\u8BA4\u6279\u51C6\u7EE7\u7EED\uFF1F") + anchor + t("\uFF1A\u53D1\u5E03 ApprovalGranted \u540E run \u6CBF\u58F0\u660E\u76F8\u4F4D\u5E8F\u5217\u7EED\u884C\u3002")) : window.confirm(t("\u786E\u8BA4\u9A73\u56DE\u7EC8\u6B62\uFF1F") + anchor + t("\uFF1A\u53D1\u5E03 ApprovalRejected \u540E run \u5C06\u8FDB\u5165\u7EC8\u6001 Failed\uFF08\u4E0D\u53EF\u6062\u590D\uFF09\u3002"));
    if (!confirmed) return;
    btn.disabled = true;
    btn.textContent = t("\u53D1\u5E03\u4E2D\u2026");
    const payload = { type: grant ? "ApprovalGranted" : "ApprovalRejected", involved: anchor };
    if (evId) payload.dedupKey = "console-" + evId;
    const r = await ctx2.api.eventPub(payload);
    if (!r.ok) {
      btn.disabled = false;
      btn.textContent = grant ? t("\u6279\u51C6\u7EE7\u7EED") : t("\u9A73\u56DE\u7EC8\u6B62");
      window.alert(t("\u4E8B\u4EF6\u53D1\u5E03\u5931\u8D25\uFF08") + String(r.error || r.status) + t("\uFF09\u2014\u2014run \u4ECD\u6302\u539F\u76F8\u4F4D\uFF0C\u672A\u88C5\u6210\u529F\u3002"));
      return;
    }
    ignored.add(key);
    const card = btn.closest(".ap-item");
    if (card) card.remove();
    ctx2.api.snapshot().then((s) => {
      if (s.ok) ctx2.store.init(s.data);
    });
  }
  async function hydrateLedger(ctx2) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const slot = document.querySelector("[data-ap-ledger]");
    if (!slot) return;
    const r = await ctx2.api.approvals(20);
    if (!el__isConnected(slot) || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
    if (!r.ok) {
      slot.innerHTML = t('<div class="small faint">\u5BA1\u6279\u8BB0\u5F55\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(r.error || r.status)) + t("\uFF09\u2014\u2014\u5224\u5B9A\u7559\u75D5\u6682\u53EA=apply_log\u3002</div>");
      return;
    }
    const rows = r.data && Array.isArray(r.data.approvals) ? r.data.approvals : [];
    if (!rows.length) {
      slot.innerHTML = t('<div class="small faint">\u5DF2\u51B3\u8BB0\u5F55\u6682\u65E0\u5224\u5B9A\u884C\uFF08\u6279\u51C6/\u9A73\u56DE\u5373\u51B3\u5373\u89C1\uFF09\u3002</div>');
      return;
    }
    slot.innerHTML = t('<h2 class="sec-title">\u5DF2\u51B3\u8BB0\u5F55</h2><div style="overflow-x:auto"><table class="tbl"><thead><tr>') + t("<th>\u65F6\u523B</th><th>\u5224\u5B9A</th><th>\u5BF9\u8C61</th><th>\u5224\u5B9A\u4E3B\u4F53</th><th>rv</th></tr></thead><tbody>") + rows.map((x) => '<tr><td class="mono small">' + esc(String(x.ts || "").replace("T", " ").slice(0, 19)) + "</td><td>" + chip(String(x.decision || ""), x.decision === "granted" ? "ok" : "esc") + '</td><td class="mono small">' + (involvedRoute2(x.obj) ? '<span class="lk" data-go="' + esc(involvedRoute2(x.obj)) + '">' + esc(String(x.obj || "")) + "</span>" : esc(String(x.obj || ""))) + '</td><td class="mono small">' + esc(String(x.judgedBy || "anonymous")) + '</td><td class="mono small">' + esc(x.rv ? "r" + x.rv : "\u2014") + "</td></tr>").join("") + "</tbody></table></div>";
  }
  function el__isConnected(x) {
    return !!(x && x.isConnected);
  }
  function inboxHtml(ctx2) {
    const inbox = ctx2.store.inbox();
    if (!inbox) return t('<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u6295\u5F71\u7F3A\u5E2D\uFF08\u5FEB\u7167\u672A\u643A\u5E26 /v1/inbox \u805A\u5408\uFF09\u2014\u2014\u5F85\u4F60\u5904\u7406\u9762\u663E\u5F0F\u7F3A\u5E2D\uFF0C\u4E0D\u88C5\u5047\u3002</div></div></div>');
    const items = Array.isArray(inbox.items) ? inbox.items : [];
    if (!items.length && !Number(inbox.total)) {
      return t('<div class="card"><div class="card-b"><div class="small faint">\u6536\u4EF6\u7BB1\u4E3A\u7A7A\uFF08\u96F6\u5F85\u529E\u2014\u2014\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div></div></div>');
    }
    return '<div class="card">' + items.slice(0, 50).map((x) => {
      const route = involvedRoute2(x && (x.link || x.ref || x.obj));
      const main = String(x && (x.title || x.summary || x.kind) || "");
      return '<div class="mini-row"><div class="mr-main"><div>' + esc(trunc(main, 80)) + " " + chip(String(x && x.class || x && x.kind || ""), "mute") + '</div><div class="mr-sub small">' + esc(trunc(String(x && (x.note || x.detail) || ""), 90)) + "</div></div>" + (route ? '<span class="ws-chip">' + chipGo(t("\u53BB\u5904\u7406"), route, "acc", true) + "</span>" : "") + "</div>";
    }).join("") + "</div>";
  }
  function approvalsScreen(el, ctx2) {
    const items = pendingItems(ctx2);
    const draftN = items.filter((x) => x.kind === "draft").length;
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u5BA1\u6279"),
      t("\u5F85\u5BA1\u533A=\u8349\u7A3F\u7F13\u5B58\uFF08DocReview \u7C7B\u540C\u6E90\uFF09\u222A ApprovalNeeded \u5F52\u6863\u884C\u3002\u4E8B\u4EF6\u884C\u6279\u51C6/\u9A73\u56DE=\u5BF9 Workflow \u951A\u53D1\u5E03 ApprovalGranted/ApprovalRejected\uFF08run \u7EED\u884C/\u7EC8\u6001 Failed\uFF09\uFF1B\u8349\u7A3F\u6279\u51C6=Apply\uFF1B\u5747=\u4EBA\u70B9\u786E\u8BA4\u3002"),
      chip(t("\u6279\u91CF\u6279\u51C6\u901A\u8DEF\u5F85\u540E\u7AEF\u8BCD\u5F62\u2014\u2014\u9010\u6761\u6279\u51C6\uFF08\u8BDA\u5B9E\u6CE8\u8BB0\uFF09"), "mute")
    ) + (items.length ? items.map(itemHtml).join("") : emptyBox(t("\u5F85\u5BA1\u533A\u4E3A\u7A7A\uFF08\u8349\u7A3F\u7F13\u5B58\u4E0E ApprovalNeeded \u5F52\u6863\u5747\u7A7A\uFF09"), "", chipGo(t("\u56DE Welcome"), "welcome", "primary", false))) + t('<div data-ap-ledger style="margin-top:20px"><div class="small faint">\u5DF2\u51B3\u8BB0\u5F55\u73B0\u8BFB\u4E2D\u2026</div></div>') + t('<h2 class="sec-title">\u6536\u4EF6\u7BB1\uFF08\u5F85\u4F60\u5904\u7406\uFF09</h2>') + inboxHtml(ctx2) + "</div>";
    items.filter((x) => x.kind === "draft").forEach((x) => hydrateDiff(ctx2, x.draft));
    hydrateLedger(ctx2);
    el.querySelectorAll("[data-ap-approve]").forEach((btn) => {
      btn.addEventListener("click", () => {
        const d = ctx2.store.drafts().find((x) => apKeyOf(x) === btn.dataset.apApprove);
        if (d) approve(ctx2, d, btn);
      });
    });
    el.querySelectorAll("[data-ap-grant]").forEach((btn) => {
      btn.addEventListener("click", () => pubGateDecision(ctx2, btn, true));
    });
    el.querySelectorAll("[data-ap-reject]").forEach((btn) => {
      btn.addEventListener("click", () => pubGateDecision(ctx2, btn, false));
    });
    el.querySelectorAll("[data-ap-ignore]").forEach((btn) => {
      btn.addEventListener("click", () => {
        ignored.add(btn.dataset.apIgnore);
        const card = btn.closest(".ap-item");
        if (card) card.remove();
      });
    });
    el.querySelectorAll("[data-go]").forEach((x) => {
      x.addEventListener("click", () => ctx2.go(x.dataset.go));
    });
  }
  function render8(el, ctx2) {
    approvalsScreen(el, ctx2);
  }

  // src/screens/misc.js
  var treeOpen = /* @__PURE__ */ new Set();
  var treeSeeded = false;
  var docFilter = { kind: "", q: "" };
  var propOpen = /* @__PURE__ */ new Set();
  var dd = { tree: null, sel: null, cache: /* @__PURE__ */ new Map() };
  function treeScreen(el, ctx2) {
    const objs = ctx2.store.objects();
    const kinds = Array.from(new Set(objs.map((o) => o.kind))).sort();
    const scopeMap = {};
    objs.forEach((o) => {
      const k = scopeMap[o.kind] = scopeMap[o.kind] || {};
      (k[o.scope] = k[o.scope] || []).push(o);
    });
    if (!treeSeeded) {
      kinds.forEach((kind) => {
        if (objs.filter((o) => o.kind === kind).length <= 3) treeOpen.add(kind);
      });
      treeSeeded = true;
    }
    const body = kinds.map((kind) => {
      const open = treeOpen.has(kind);
      const n = objs.filter((o) => o.kind === kind).length;
      const rows = open ? Object.keys(scopeMap[kind] || {}).sort().map((scope) => {
        const items = (scopeMap[kind][scope] || []).slice().sort((a, b) => a.name < b.name ? -1 : 1);
        return '<div class="tree-scope"><span class="small faint mono">' + esc(scope) + "/</span>" + items.map((o) => '<div class="tree-row leaf" data-go="' + esc(objRoute(o.kind, o.scope, o.name)) + '"><span style="flex:1;min-width:0">' + esc(o.name) + '</span> <span class="faint tiny mono">r' + esc(String(o.rv || 0)) + "</span>" + (o.kind === "ServiceDesc" ? '<span class="ws-chip">' + chipGo(t("Service \u5927\u5C4F"), "service/" + encodeURIComponent(String(o.name || "")), "acc", true, t("\u6DF1\u94FE Service \u5927\u5C4F")) + "</span>" : "") + "</div>").join("") + "</div>";
      }).join("") : "";
      return '<div class="tree-node"><div class="tree-row grp" data-tree-toggle="' + esc(kind) + '"><b>' + esc(kind) + "</b>" + chip(n + t(" \u4E2A"), n ? "acc" : "mute") + '<span class="faint">' + (open ? "\u25BE" : "\u25B8") + "</span></div>" + rows + "</div>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u7CFB\u7EDF\u6811"), t("kinds \u2192 scope \u2192 \u5BF9\u8C61\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF1B\u5BF9\u8C61\u540D\u53EF\u70B9\u8FDB\u5DE5\u4F5C\u53F0\uFF1B\u22643 \u5BF9\u8C61\u7684 kind \u9ED8\u8BA4\u5C55\u5F00\uFF09")) + (kinds.length ? '<div class="tree-box" style="min-width:min(680px,100%)">' + body + "</div>" : emptyBox(t("\u5FEB\u7167\u6682\u65E0\u5BF9\u8C61\u679A\u4E3E"), t("\u7B49\u5F85\u9996\u4E2A\u5FEB\u7167\u805A\u5408\u3002"))) + "</div>";
    el.querySelectorAll("[data-tree-toggle]").forEach((row) => {
      row.addEventListener("click", () => {
        const k = row.dataset.treeToggle;
        if (treeOpen.has(k)) treeOpen.delete(k);
        else treeOpen.add(k);
        treeScreen(el, ctx2);
      });
    });
  }
  var DOC_KINDS = ["Goal", "Rule", "Runbook", "DesignDoc", "DriftReport", "DocReview", "Agent", "WorkService"];
  function fmMeta(content) {
    if (typeof content !== "string" || !content.slice(0, 4e3).includes("tags:")) return null;
    const head = content.slice(0, 4e3);
    const tags = [];
    let importance = null;
    const flat = head.match(/tags:\s*\[([^\]]*)\]/);
    if (flat) flat[1].split(",").forEach((x) => {
      const v = x.trim().replace(/^['"]|['"]$/g, "");
      if (v) tags.push(v);
    });
    const imp = head.match(/^importance:\s*(\d{1,3})\s*$/m);
    if (imp) importance = Number(imp[1]);
    return { tags, importance };
  }
  function docsScreen(el, ctx2) {
    let objs = ctx2.store.objects();
    const presentKinds = Array.from(new Set(objs.map((o) => o.kind)));
    const docKinds = DOC_KINDS.filter((k) => presentKinds.includes(k));
    const pool = docKinds.length ? objs.filter((o) => docKinds.includes(o.kind)) : objs;
    let rows = pool;
    if (docFilter.kind) rows = rows.filter((o) => o.kind === docFilter.kind);
    if (docFilter.q) {
      const q = docFilter.q.toLowerCase();
      rows = rows.filter((o) => (o.kind + "/" + o.scope + "/" + o.name).toLowerCase().includes(q));
    }
    if (docFilter.tag) {
      const tg = docFilter.tag.toLowerCase();
      rows = rows.filter((o) => {
        const c = ctx2.store.cachedDoc(o.kind, o.scope || "global", o.name, o.rv);
        const meta = c ? fmMeta(c.content) : null;
        return !!(meta && meta.tags.some((x) => x.toLowerCase().includes(tg)));
      });
    }
    const draftSet = new Set(ctx2.store.drafts().map((d) => d.kind + "|" + (d.scope || "global") + "|" + d.name));
    el.innerHTML = '<div class="view-inner">' + screenHead(
      t("\u6587\u6863\u5E93"),
      t("\u76EE\u5F55\u6811\u5F0F\u6587\u6863\u6E05\u5355\u2014\u2014tag \u8FC7\u6EE4+\u7248\u672C/\u91CD\u8981\u5EA6\u5217\uFF1B\u70B9\u884C\u8FDB\u6E32\u67D3\u89C6\u56FE\u4E0E AI \u4FEE\u6587\u6863"),
      docKinds.length ? "" : chip(t("\u672A\u6CE8\u518C\u58F0\u660E kind\u2014\u2014\u5168\u91CF\u679A\u4E3E\u515C\u5E95"), "mute")
    ) + '<div class="card" style="margin-bottom:10px"><div class="card-b small">' + esc(t("\u65B0\u5EFA\u6587\u6863\uFF1A\u5BF9\u8BDD\u63D0\u4EA4\u9700\u6C42\uFF08\u603B\u89C8\u6216\u4F1A\u8BDD\u5C4F\uFF09\u2192 agent \u8D77\u8349 \u2192 Apply \u843D\u5E93\u2014\u2014\u6587\u6863\u521B\u5EFA\u8D70\u5BF9\u8BDD+Apply\uFF0C\u65E0\u72EC\u7ACB\u8868\u5355\u3002")) + " " + chipGo(t("\u6253\u5F00\u5BF9\u8BDD"), "chat/core-headless", "acc", true, t("\u5728\u4F1A\u8BDD\u5C4F\u63D0\u4EA4\u9700\u6C42\u2014\u2014agent \u8D77\u8349\u540E Apply \u843D\u5E93")) + '</div></div><div class="filter-bar">' + t('<input class="inp" id="docs-q" placeholder="\u6309 kind/scope/name \u8FC7\u6EE4\u2026" value="') + esc(docFilter.q) + '">' + t('<input class="inp" id="docs-tag" placeholder="tag \u8FC7\u6EE4\uFF08\u7F13\u5B58\u6B63\u6587\u73B0\u5339\u914D\uFF09\u2026" value="') + esc(docFilter.tag || "") + '" style="max-width:220px"><span class="fchip' + (!docFilter.kind ? " on" : "") + t('" data-docs-kind="">\u5168\u90E8</span>') + docKinds.map((k) => '<span class="fchip' + (docFilter.kind === k ? " on" : "") + '" data-docs-kind="' + esc(k) + '">' + esc(k) + "</span>").join("") + "</div>" + (rows.length ? '<div class="card">' + rows.slice(0, 300).map((o) => {
      const c = ctx2.store.cachedDoc(o.kind, o.scope || "global", o.name, o.rv);
      const meta = c ? fmMeta(c.content) : null;
      const tagChips = (meta && meta.tags ? meta.tags.slice(0, 4) : []).map((x) => chip(x, "mute")).join("");
      const imp = meta && meta.importance != null ? chip("I:" + meta.importance, meta.importance >= 80 ? "acc" : "mute") : chip(t("\u91CD\u8981\u5EA6\u2014"), "mute");
      const sync = syncBadge(ctx2.store.runs(), o.name);
      return '<div class="mini-row" data-go="' + esc(docRoute(o.kind, o.scope, o.name)) + '"><div class="mr-main"><div><b>' + esc(o.name) + "</b>" + (draftSet.has(o.kind + "|" + o.scope + "|" + o.name) ? '<span class="draft-dot">\u25CF</span>' : "") + '</div><div class="mr-sub small mono">' + esc(o.kind + "/" + o.scope) + " \xB7 rv " + esc(String(o.rv != null ? o.rv : "\u2014")) + '</div><div class="mr-sub">' + tagChips + imp + sync + '</div></div><span class="ws-chip">' + chipGo(t("\u6253\u5F00"), docRoute(o.kind, o.scope, o.name), "acc", true, t("\u6E32\u67D3\u89C6\u56FE")) + "</span></div>";
    }).join("") + (rows.length > 300 ? t('<div class="small faint" style="padding:8px 14px">\u2026\u5171 ') + rows.length + t(" \u884C\u2014\u2014\u7528\u8FC7\u6EE4\u6536\u7A84\u3002</div>") : "") + "</div>" : emptyBox(t("\u7A97\u53E3\u5185\u65E0\u6587\u6863"), t("\u8FC7\u6EE4\u96F6\u547D\u4E2D\u6216\u5FEB\u7167\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002"))) + '<div class="small faint" style="margin-top:8px">' + esc(t("tag/\u91CD\u8981\u5EA6\u6765\u81EA\u7F13\u5B58\u6B63\u6587\u7684 frontmatter\uFF08\u9884\u53D6\u7A97\u53E3 24 \u7BC7\u2014\u2014\u672A\u7F13\u5B58\u7BC7\u663E\u793A \u2014\uFF0C\u73B0\u5F00\u5373\u8BFB\uFF09\u3002")) + "</div></div>";
    const inp = el.querySelector("#docs-q");
    if (inp) inp.addEventListener("input", () => {
      docFilter.q = inp.value;
      docsScreen(el, ctx2);
    });
    const tagInp = el.querySelector("#docs-tag");
    if (tagInp) tagInp.addEventListener("input", () => {
      docFilter.tag = tagInp.value;
      docsScreen(el, ctx2);
    });
    el.querySelectorAll("[data-docs-kind]").forEach((c) => {
      c.addEventListener("click", () => {
        docFilter.kind = c.dataset.docsKind;
        docsScreen(el, ctx2);
      });
    });
    el.querySelectorAll(".mini-row[data-go]").forEach((row) => {
      row.addEventListener("click", () => ctx2.go(row.dataset.go));
    });
  }
  async function hydrateDesignDoc(el, ctx2, repo, path) {
    const stamp = typeof location !== "undefined" ? location.hash : "";
    const cacheKey = repo + "|" + path;
    if (!dd.cache.has(cacheKey)) {
      const r = await ctx2.api.designDoc(repo, path);
      if (!el.isConnected || (typeof location !== "undefined" ? location.hash : "") !== stamp) return;
      dd.cache.set(cacheKey, r);
    }
    const res = dd.cache.get(cacheKey);
    const box = document.querySelector("[data-dd-body]");
    if (!box) return;
    if (!res || !res.ok) {
      box.innerHTML = t('<div class="ob-empty">\u6B63\u6587\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(res && res.error || res && res.status || t("\u7F51\u7EDC"))) + t("\uFF09\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF0C\u91CD\u8FDB\u5373\u73B0\u8BFB\u3002</div>");
      return;
    }
    const data = res.data && typeof res.data === "object" ? res.data : {};
    const content = data.content != null ? String(data.content) : typeof res.data === "string" ? res.data : "";
    box.innerHTML = '<div data-md-render="1">' + renderMarkdown(content) + "</div>" + rawDisclosure(t("\u539F\u6587\uFF08HEAD\uFF09"), content) + '<div class="small faint" style="margin-top:6px">' + (data.sha ? "HEAD " + esc(String(data.sha).slice(0, 8)) : "HEAD") + t("\uFF08git \u6743\u5A01\u96F6\u7F13\u5B58\uFF09 \xB7 \u8349\u7A3F/Apply \u901A\u8DEF\u7ECF\u7F16\u8F91\u5668\u5C4F\uFF08doc/\u2026\uFF09\u3002</div>");
  }
  function designDocScreen(el, ctx2) {
    const repo = ctx2.params && ctx2.params.repo;
    const path = ctx2.params && ctx2.params.path;
    if (repo && path) {
      dd.sel = { repo, path };
    }
    const sel = dd.sel;
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u8BBE\u8BA1\u6587\u6863"), t("\u767D\u540D\u5355\u4ED3\u6811\uFF08GET /v1/design-docs\u2014\u2014HEAD \u6743\u5A01\u96F6\u7F13\u5B58\uFF09")) + t('<div class="dd-grid"><div data-dd-tree><div class="small faint">\u4ED3\u6811\u73B0\u8BFB\u4E2D\u2026</div></div>') + '<div class="card"><div class="card-b" data-dd-body>' + (sel ? t('<div class="small faint">\u6B63\u6587\u73B0\u8BFB\u4E2D\u2026</div>') : emptyBox(t("\u672A\u9009\u6863"), t("\u5DE6\u4FA7\u9009\u62E9\u4ED3\u6587\u4EF6\uFF08\u9ED8\u8BA4\u9996\u4ED3\u9996\u6587\u4EF6\uFF09\u3002"))) + "</div></div></div></div>";
    ctx2.api.designDocs().then((r) => {
      if (!el.isConnected) return;
      const treeBox = el.querySelector("[data-dd-tree]");
      if (!treeBox) return;
      if (!r.ok) {
        treeBox.innerHTML = t('<div class="ob-empty">\u4ED3\u6811\u6682\u4E0D\u53EF\u8BFB\uFF08') + esc(String(r.error || r.status)) + "\uFF09\u3002</div>";
        return;
      }
      dd.tree = r.data;
      const files = Array.isArray(r.data) ? r.data : r.data && r.data.files || [];
      if (!files.length) {
        treeBox.innerHTML = t('<div class="ob-empty">\u4ED3\u6811\u4E3A\u7A7A\uFF08\u8BDA\u5B9E\u7A7A\u6001\uFF09\u3002</div>');
        return;
      }
      treeBox.innerHTML = files.map((f) => {
        const fr = f.repo || (r.data.repo || "");
        const fp = f.path || String(f);
        const key = fr + "|" + fp;
        return '<div class="mini-row' + (sel && sel.repo === fr && sel.path === fp ? " sel" : "") + '" data-dd-file="' + esc(key) + '"><div class="mr-main"><div class="mono small">' + esc(trunc(fp, 40)) + '</div><div class="mr-sub tiny">' + esc(fr) + "</div></div></div>";
      }).join("");
      treeBox.querySelectorAll("[data-dd-file]").forEach((row) => {
        row.addEventListener("click", () => {
          const [fr, ...rest] = row.dataset.ddFile.split("|");
          const fp = rest.join("|");
          ctx2.go("designdoc/" + encodeURIComponent(fr) + "/" + encodeURIComponent(fp));
        });
      });
      if (!sel && files.length) {
        const f0 = files[0];
        ctx2.go("designdoc/" + encodeURIComponent(String(f0.repo || r.data.repo || "")) + "/" + encodeURIComponent(String(f0.path || f0)));
        return;
      }
      if (sel) hydrateDesignDoc(el, ctx2, sel.repo, sel.path);
    }).catch(() => {
    });
  }
  function connectorsScreen(el, ctx2) {
    const kinds = ["ConnectorOps", "WorkService", "Provider", "Ref"];
    const secs = kinds.map((kind) => {
      const rows = ctx2.store.refs(kind);
      return '<section class="card"><div class="card-h"><span class="h-title">' + esc(kind) + '</span><span class="h-right">' + chip(rows.length + t(" \u4E2A"), rows.length ? "acc" : "mute") + '</span></div><div class="card-b">' + (rows.length ? rows.map((o) => '<div class="mini-row" data-go="' + esc(docRoute(kind, o.scope, o.name)) + '"><div class="mr-main"><div><b class="mono">' + esc(o.name) + '</b></div><div class="mr-sub small mono">' + esc(o.scope) + " \xB7 r" + esc(String(o.rv || 0)) + '</div></div><span class="ws-chip">' + chipGo(t("\u5DE5\u4F5C\u53F0"), objRoute(kind, o.scope, o.name), "mute", true) + "</span></div>").join("") : t('<div class="small faint">\u8BE5 kind \u6682\u65E0\u5728\u518C\u5BF9\u8C61\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002</div>')) + "</div></section>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u8FDE\u63A5\u5C42"), t("ConnectorOps / WorkService / Provider / Ref \u56DB\u65CF\uFF08\u6587\u6863\u4E0E\u8D44\u6E90\u540C\u89C4\u5219\uFF09")) + '<div class="grid cols-2">' + secs + "</div></div>";
    el.querySelectorAll(".mini-row[data-go]").forEach((row) => {
      row.addEventListener("click", () => ctx2.go(row.dataset.go));
    });
  }
  function proposalsScreen(el, ctx2) {
    const rows = ctx2.store.proposals();
    const cards = rows.map((p, i) => {
      const keys = p && typeof p === "object" ? Object.keys(p) : [];
      const key = String(p && p.name || "") || "#" + i;
      const open = propOpen.has(key);
      const known = ["title", "summary", "status", "proposer", "createdAt", "scope", "name"];
      const sub = known.filter((k) => p[k] != null && k !== "title").map((k) => esc(k) + ": " + esc(trunc(String(p[k]), 40))).join(" \xB7 ");
      const extras = keys.filter((k) => !known.includes(k) && k !== "summary" && p[k] != null && p[k] !== "").map((k) => '<span class="k">' + esc(k) + '</span><span class="v">' + esc(trunc(String(p[k]), 200)) + "</span>").join("");
      return '<div class="mini-row" data-prop="' + esc(key) + '"><div class="mr-main"><div><b>' + esc(String(p.title || p.name || t("\u672A\u547D\u540D\u63D0\u6848"))) + "</b> " + chip(String(p.status || t("\u5F85\u5B9A")), "mute") + '</div><div class="mr-sub small">' + (sub || t('<span class="faint">\u8981\u7D20\u5B57\u6BB5\u7F3A\u5E2D\u2014\u2014\u4E0D\u7F16\u9020</span>')) + "</div>" + (open ? '<div class="mr-sub" style="margin-top:6px;white-space:pre-wrap">' + esc(trunc(String(p.summary || ""), 2e3) || t("summary \u7F3A\u5E2D\u2014\u2014\u4E0D\u7F16\u9020")) + "</div>" + (extras ? '<div class="kv" style="margin-top:6px">' + extras + "</div>" : "") + '<div class="small faint" style="margin-top:6px">' + esc(t("\u518D\u70B9\u4E00\u6B21\u5361\u7247\u6536\u8D77\u3002")) + "</div>" : "") + '</div><span class="ws-chip">' + chipGo(t("\u53BB\u5BA1\u6279"), "approvals", "acc", true) + "</span></div>";
    }).join("");
    el.innerHTML = '<div class="view-inner">' + screenHead(t("\u9700\u6C42\u63D0\u6848"), t("\u63D0\u6848\u9762\uFF08\u70B9\u5361\u7247\u5C55\u5F00 summary \u5168\u6587\u4E0E\u8981\u7D20\uFF09\u2014\u2014\u6279\u51C6/\u9A73\u56DE\u8D70\u5BA1\u6279\u4E0E Apply\uFF08LLM \u65E0 Apply \u6743\uFF09")) + (rows.length ? '<div class="card">' + cards + "</div>" : emptyBox(t("\u6682\u65E0\u63D0\u6848"), t("\u5FEB\u7167 proposals \u805A\u5408\u4E3A\u7A7A\uFF08\u7A7A\u6001\u663E\u5F0F\uFF09\u3002"))) + "</div>";
    el.querySelectorAll("[data-prop]").forEach((row) => {
      row.addEventListener("click", (ev) => {
        if (ev.target.closest("[data-go]")) return;
        const key = row.dataset.prop;
        if (propOpen.has(key)) propOpen.delete(key);
        else propOpen.add(key);
        proposalsScreen(el, ctx2);
      });
    });
  }
  function systemScreen(el, ctx2) {
    const scope = String(ctx2.params.name || "");
    const objs = ctx2.store.objects().filter((o) => o.scope === scope);
    const evs = ctx2.store.events().filter((e) => String(e.involved || "").includes(scope)).slice(0, 10);
    const byKind = {};
    objs.forEach((o) => {
      byKind[o.kind] = (byKind[o.kind] || 0) + 1;
    });
    el.innerHTML = '<div class="view-inner">' + screenHead(scope, "scope=" + scope + t(" \u7684\u5BF9\u8C61\u4E0E\u4E8B\u4EF6\uFF08\u5FEB\u7167\u805A\u5408\u6295\u5F71\uFF09"), chipGo(t("\u7CFB\u7EDF\u6811"), "tree", "mute", false)) + '<div class="statbar">' + Object.keys(byKind).sort().map((k) => '<div class="stat" data-go="' + esc("obj/" + encodeURIComponent(k)) + '" role="link"><div class="st-num">' + byKind[k] + '</div><div class="st-label">' + esc(k) + "</div></div>").join("") + "</div>" + (objs.length ? t('<h2 class="sec-title">\u5BF9\u8C61</h2><div class="card">') + objs.map((o) => '<div class="mini-row" data-go="' + esc(objRoute(o.kind, o.scope, o.name)) + '"><div class="mr-main"><div><b>' + esc(o.name) + "</b> " + kindBadge(o.kind) + '</div><div class="mr-sub small mono">r' + esc(String(o.rv || 0)) + "</div></div></div>").join("") + "</div>" : emptyBox(t("\u8BE5 scope \u6682\u65E0\u5728\u518C\u5BF9\u8C61"), t("\u4E0D\u88C5\u5047\u2014\u2014scope \u8BCD\u5F62\u53EF\u80FD\u4E3A\u5B9E\u4F8B\u6876\uFF08\u6811\u5185\u53EF\u89C1\uFF09\u3002"))) + (evs.length ? t('<h2 class="sec-title">\u6700\u8FD1\u76F8\u5173\u4E8B\u4EF6</h2><div class="card">') + evs.map((e) => '<div class="mini-row" data-go="event/' + encodeURIComponent(String(e.id ?? "")) + '"><div class="mr-main"><div>' + esc(String(e.type || "")) + '</div><div class="mr-sub small mono">' + esc(trunc(String(e.involved || "\u2014"), 60)) + "</div></div></div>").join("") + "</div>" : "") + "</div>";
    el.querySelectorAll("[data-go]").forEach((x) => {
      x.addEventListener("click", () => ctx2.go(x.dataset.go));
    });
  }
  function render9(el, ctx2) {
    const r = ctx2.route;
    if (r === "tree") treeScreen(el, ctx2);
    else if (r === "docs") docsScreen(el, ctx2);
    else if (r === "designdoc") designDocScreen(el, ctx2);
    else if (r === "connectors") connectorsScreen(el, ctx2);
    else if (r === "proposals") proposalsScreen(el, ctx2);
    else if (r === "s") systemScreen(el, ctx2);
    else el.innerHTML = '<div class="view-inner">' + emptyBox(t("\u672A\u77E5\u8584\u5C4F"), String(r)) + "</div>";
  }
  function renderBiz(el, ctx2) {
    el.innerHTML = '<div class="view-inner"><div class="biz-hero"><div class="biz-hero-mark">B</div><h2 class="sec-title">' + esc(t("\u4E1A\u52A1\u63A7\u5236\u53F0")) + '</h2><div class="small faint">' + esc(t("\u4E1A\u52A1\u89C6\u89D2\u5360\u4F4D\u2014\u2014\u540E\u7EED ServiceAgent \u767B\u5F55\u540E\u5728\u6B64\u627F\u8F7D\u4E1A\u52A1\u5927\u5C4F\u4E0E\u64CD\u4F5C\u9762\uFF08\u5F53\u524D\u96F6\u5185\u5BB9\uFF0C\u663E\u5F0F\u5360\u4F4D\u4E0D\u88C5\u5047\uFF09\u3002")) + "</div></div></div>";
  }

  // src/main.js
  var BASES = typeof window !== "undefined" && window.__CORE_CONSOLE_BASES__ || {};
  var store = createStore();
  var api = createApi(BASES.api || "", { apply: BASES.apply, drafts: BASES.drafts });
  var router = createRouter({ defaultRoute: "welcome" });
  var ctx = { store, api, router, go: (p) => router.go(p), params: {}, route: "" };
  var NAV_TREE = [
    {
      level: "sys",
      label: "\u7CFB\u7EDF\u63A7\u5236\u53F0",
      icon: "\u25A3",
      items: [
        ["welcome", "\u603B\u89C8", "\u2302"],
        ["report", "\u5173\u6CE8\u62A5\u544A", "\u2248"],
        ["tasks", "\u4EFB\u52A1\u4E2D\u5FC3", "\u229E"],
        ["workflows", "Workflows", "\u27F3"],
        ["events", "\u4E8B\u4EF6", "\u26A1"],
        ["approvals", "\u5BA1\u6279", "\u2713"],
        ["proposals", "\u9700\u6C42\u63D0\u6848", "\u2726"]
      ],
      sub: { label: "\u57FA\u7840\u8BBE\u65BD", items: [
        ["chat/core-headless", "\u4F1A\u8BDD", "\u{1F4AC}"],
        ["tree", "\u7CFB\u7EDF\u6811", "\u2443"],
        ["docs", "\u6587\u6863\u5E93", "\u25A4"],
        ["obj", "\u5BF9\u8C61\u5DE5\u4F5C\u53F0", "\u25AD"],
        ["connectors", "\u8FDE\u63A5", "\u2301"]
      ] }
    },
    {
      level: "biz",
      label: "\u4E1A\u52A1\u63A7\u5236\u53F0",
      icon: "\u25C6",
      items: [["biz", "\u4E1A\u52A1\u603B\u89C8", "\u2302"]],
      sub: null
    }
  ];
  var curLevel = "sys";
  try {
    curLevel = typeof localStorage !== "undefined" && localStorage.getItem("cc.level") || "sys";
  } catch {
  }
  function langToggleHtml() {
    const cur = getLocale();
    const btn = (id, label) => '<button class="lang-btn' + (cur === id ? " active" : "") + '" data-lang="' + id + '">' + label + "</button>";
    return '<span class="lang-toggle">' + btn("zh", "\u4E2D\u6587") + btn("en", "EN") + "</span>";
  }
  var foldState = { "\u57FA\u7840\u8BBE\u65BD": false };
  function approvalBadge() {
    try {
      const n = (store.drafts() || []).length + (store.events() || []).filter((e) => e && e.type === "ApprovalNeeded").length;
      return n > 0 ? '<span class="nav-badge">' + n + "</span>" : "";
    } catch {
      return "";
    }
  }
  function navTree() {
    return NAV_TREE;
  }
  function sidebarHtml() {
    const sections = navTree().map((sec) => {
      const open = curLevel === sec.level;
      const items = sec.items.map(([r, label, icon]) => {
        const badge = r === "approvals" ? approvalBadge() : "";
        return '<div class="nav-item l2" data-nav="' + esc(r.split("/")[0]) + '" data-go="' + esc(r) + '"><span class="nav-ico">' + icon + "</span>" + esc(t(label)) + badge + "</div>";
      }).join("");
      const sub = sec.sub ? '<div class="nav-sub-h" data-fold="' + esc(sec.sub.label) + '"><span class="nav-ico fold-chev">' + (foldState[sec.sub.label] ? "\u25B8" : "\u25BE") + "</span>" + esc(t(sec.sub.label)) + "</div>" + (foldState[sec.sub.label] ? "" : '<nav class="side-nav l3">' + sec.sub.items.map(([r, label, icon]) => '<div class="nav-item l3" data-nav="' + esc(r.split("/")[0]) + '" data-go="' + esc(r) + '"><span class="nav-ico">' + icon + "</span>" + esc(t(label)) + "</div>").join("") + "</nav>") : "";
      return '<div class="nav-section' + (open ? " open" : "") + '" data-section="' + sec.level + '"><div class="nav-section-h" data-level="' + sec.level + '"><span class="nav-ico">' + sec.icon + "</span>" + esc(t(sec.label)) + '<span class="fold-chev">' + (open ? "\u25BE" : "\u25B8") + "</span></div>" + (open ? '<nav class="side-nav l2">' + items + "</nav>" + sub : "") + "</div>";
    }).join("");
    return sections;
  }
  function sideFootHtml() {
    return '<div class="side-foot"><span class="lamp off" data-conn-lamp></span><span class="small faint" data-conn-label>' + esc(t("\u8FDE\u63A5\u4E2D\u2026")) + "</span></div>";
  }
  function crumbLabel(route) {
    const bare = String(route || "").split("/")[0];
    for (const sec of NAV_TREE) {
      for (const [r, label] of sec.items) if (r.split("/")[0] === bare) return t(label);
      if (sec.sub) {
        for (const [r, label] of sec.sub.items) if (r.split("/")[0] === bare) return t(label);
      }
    }
    const extra = { task: t("\u4EFB\u52A1\u8BE6\u60C5"), event: t("\u4E8B\u4EF6\u8BE6\u60C5"), doc: t("\u6587\u6863\u7F16\u8F91\u5668"), designdoc: t("\u8BBE\u8BA1\u6587\u6863"), wf: t("\u8FD0\u884C\u8BE6\u60C5"), session: t("\u4F1A\u8BDD"), service: t("Service \u5927\u5C4F"), s: t("\u7CFB\u7EDF\u8BE6\u60C5"), inbox: t("\u6536\u4EF6\u7BB1") };
    return extra[bare] || bare;
  }
  function shellHtml() {
    return '<div class="app"><aside class="sidebar"><div class="side-logo"><span class="logo-mark">C</span><span><span class="logo-name">Core Console</span><br><span class="logo-sub">' + esc(t("web \u91CD\u6784\u8D77\u70B9")) + "</span></span></div><div data-nav-tree>" + sidebarHtml() + "</div>" + sideFootHtml() + '</aside><div class="main-wrap"><header class="topbar"><span class="topbar-crumb" data-topbar-crumb></span><span class="topbar-right"><span class="lamp off" data-conn-lamp-top></span><span class="small faint" data-conn-label-top>' + esc(t("\u8FDE\u63A5\u4E2D\u2026")) + "</span>" + langToggleHtml() + '</span></header><main class="main"><div class="view" id="view"></div></main></div></div>';
  }
  function paintNav(route) {
    document.querySelectorAll("[data-nav]").forEach((n) => {
      const r = String(n.dataset.nav || "");
      const on = route === r || r === "tasks" && route === "task" || r === "events" && route === "event" || r === "workflows" && route === "wf" || r === "docs" && (route === "doc" || route === "designdoc") || r === "welcome" && route === "service" || r === "chat" && (route === "session" || route.startsWith("chat"));
      n.classList.toggle("active", on);
      const section = n.closest(".nav-section");
      if (on && section && section.dataset.section && section.dataset.section !== curLevel) {
        curLevel = section.dataset.section;
        try {
          localStorage.setItem("cc.level", curLevel);
        } catch {
        }
        const tree = document.querySelector("[data-nav-tree]");
        if (tree) tree.innerHTML = sidebarHtml();
      }
    });
  }
  function screenOf(fn) {
    return (el, params) => {
      ctx.params = params;
      ctx.route = router.current.route;
      fn(el, ctx);
      paintNav(router.current.route);
      try {
        const crumb = document.querySelector("[data-topbar-crumb]");
        if (crumb) crumb.textContent = crumbLabel(router.current.route);
      } catch {
      }
    };
  }
  var lastConnState = "connecting";
  function paintConn(state) {
    lastConnState = state;
    const text2 = { connected: t("\u5B9E\u65F6\uFF08WS patch\uFF09"), reconnecting: t("\u91CD\u8FDE\u4E2D\u2026"), degraded: t("\u964D\u7EA7\u8F6E\u8BE2\uFF0830s\uFF09"), connecting: t("\u8FDE\u63A5\u4E2D\u2026") }[state] || state;
    const cls = { connected: "ok", reconnecting: "warn", degraded: "warn", connecting: "off" }[state] || "off";
    [["[data-conn-lamp]", "[data-conn-label]"], ["[data-conn-lamp-top]", "[data-conn-label-top]"]].forEach(([ls, ltx]) => {
      document.querySelectorAll(ls).forEach((lamp) => {
        lamp.className = "lamp " + cls + (state === "connected" ? " pulse-ok" : "");
      });
      document.querySelectorAll(ltx).forEach((label) => {
        label.textContent = text2;
      });
    });
  }
  var hydrating = false;
  async function hydrateDocs() {
    if (hydrating) return;
    hydrating = true;
    try {
      const stale = store.staleDocs(24);
      let i = 0;
      const worker = async () => {
        while (i < stale.length) {
          const r = stale[i++];
          const res = await api.object(r.kind, r.scope, r.name);
          if (res.ok) {
            const data = res.data && typeof res.data === "object" ? res.data : {};
            const content = data.content != null ? String(data.content) : typeof res.data === "string" ? res.data : JSON.stringify(res.data, null, 2);
            store.cacheDoc(r.kind, r.scope, r.name, r.rv, content);
          }
        }
      };
      await Promise.all(Array.from({ length: Math.min(4, stale.length) }, worker));
    } finally {
      hydrating = false;
    }
  }
  function mount(root, opts = {}) {
    if (!root) return null;
    root.classList.add("cc-app");
    root.innerHTML = shellHtml();
    const view = root.querySelector("#view");
    store.subscribe(() => {
      if (view.isConnected) router.render();
    });
    root.addEventListener("click", (ev) => {
      const lv = ev.target.closest("[data-level]");
      if (lv) {
        ev.preventDefault();
        curLevel = lv.dataset.level;
        try {
          localStorage.setItem("cc.level", curLevel);
        } catch {
        }
        const tree = root.querySelector("[data-nav-tree]");
        if (tree) tree.innerHTML = sidebarHtml();
        router.go(curLevel === "biz" ? "biz" : "welcome");
        return;
      }
      const fold = ev.target.closest("[data-fold]");
      if (fold) {
        ev.preventDefault();
        const g = fold.dataset.fold;
        foldState[g] = !foldState[g];
        const tree = root.querySelector("[data-nav-tree]");
        if (tree) tree.innerHTML = sidebarHtml();
        return;
      }
      const langBtn = ev.target.closest("[data-lang]");
      if (langBtn) {
        ev.preventDefault();
        setLocale(langBtn.dataset.lang);
        return;
      }
      const t2 = ev.target.closest("[data-go]");
      if (t2) {
        ev.preventDefault();
        router.go(t2.dataset.go);
      }
    });
    onLocaleChange(() => {
      const tree = root.querySelector("[data-nav-tree]");
      if (tree) tree.innerHTML = sidebarHtml();
      paintConn(lastConnState);
      if (view.isConnected) router.render();
      try {
        document.title = getLocale() === "en" ? "Core Console v2" : "Core Console v2\uFF08web \u91CD\u6784\u8F7D\u4F53\xB7\u5E76\u884C\u9A8C\u6536\u9762\uFF09";
      } catch {
      }
    });
    router.route("welcome", screenOf(render2)).route("obj", screenOf(render5)).route("obj/:kind", screenOf(render5)).route("obj/:kind/:scope/:name", screenOf(render5)).route("events", screenOf(render4)).route("event/:id", screenOf(render4)).route("tasks", screenOf(render6)).route("task/:id", screenOf(render6)).route("report", screenOf(renderReport)).route("doc/:name", screenOf(render7)).route("doc/:kind/:scope/:name", screenOf(render7)).route("chat/:agent", screenOf(render)).route("session/:id", screenOf(render)).route("approvals", screenOf(render8)).route("inbox", screenOf(render8)).route("proposals", screenOf(render9)).route("workflows", screenOf(render3)).route("wf/:name", screenOf(render3)).route("wf/:name/r/:n", screenOf(render3)).route("tree", screenOf(render9)).route("docs", screenOf(render9)).route("designdoc", screenOf(render9)).route("designdoc/:repo/:path", screenOf(render9)).route("connectors", screenOf(render9)).route("s/:name", screenOf(render9)).route("service/:name", screenOf(renderService)).route("biz", screenOf(renderBiz));
    router.mount(view);
    const stream = connectStream({
      url: BASES.ws || wsUrlOf(opts.location || (typeof location !== "undefined" ? location : null)),
      base: BASES.api || "",
      onSnapshot: (s) => {
        store.setConn("connected");
        store.init(s);
        hydrateDocs();
      },
      onPatch: (patch, msg) => {
        if (!store.applyPatch(patch, msg)) {
          api.snapshot().then((r) => {
            if (r.ok) {
              store.init(r.data);
              hydrateDocs();
            }
          });
        }
      },
      onState: (s) => {
        paintConn(s);
      }
    });
    return { store, api, router, stream };
  }
  if (typeof document !== "undefined") {
    const boot = () => {
      const root = document.querySelector("[data-core-console]") || document.getElementById("app");
      if (root && !root.classList.contains("cc-app")) mount(root);
    };
    if (document.readyState === "loading") document.addEventListener("DOMContentLoaded", boot);
    else boot();
  }
  if (typeof window !== "undefined") {
    window.CoreConsoleV2 = { mount, store, api, router, connectStream, wsUrlOf };
  }
  return __toCommonJS(main_exports);
})();
