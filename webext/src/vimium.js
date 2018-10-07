// The code in this file was copied and adapted from vimium's codebase.
// It's MIT licensed, so there should be no problem mixing it with
// Browsh's codebase or even relicensing the code as LGPLv2.
//
// Here is the list of changes made to vimium's original code in order
// to better fullfill the needs of Browsh:
//
// In line 11 of getLocalHints the following code was added:
// 
//          if (requireHref && element.href && visibleElement.length >  0) {
//            visibleElement[0]["href"] = element.href;
//          }
//
// Also in getLocalHints the following outcommented code was replaced
// with the code preceding it in order to prevent Firefox from crashing:
//
//					try {
//					var testextend = extend(visibleElement, { rect: rects[0] });
//					nonOverlappingElements.push(testextend);
//					} catch(error) {
//						nonOverlappingElements.push(visibleElement);
//					}
//         /*nonOverlappingElements.push(extend(visibleElement, {
//            rect: rects[0]
//          }));*/
//
// The following lines just before the return statement in getLocalHints were
// commented out, because we're currently not using this functionality and the
// settings for it simply don't exist in Browsh (yet).
// 
//      /*if (Settings.get("filterLinkHints")) {
//        for (m = 0, len3 = localHints.length; m < len3; m++) {
//          hint = localHints[m];
//          extend(hint, this.generateLinkText(hint));
//        }
//      }*/
//
// In multiple places Utils.isFirefox() calls were commented out or replaced with true,
// because Browsh is currently assuming that it is being run in Firefox.

var Rect = {
    create: function(x1, y1, x2, y2) {
      return {
        bottom: y2,
        top: y1,
        left: x1,
        right: x2,
        width: x2 - x1,
        height: y2 - y1
      };
    },
    copy: function(rect) {
      return {
        bottom: rect.bottom,
        top: rect.top,
        left: rect.left,
        right: rect.right,
        width: rect.width,
        height: rect.height
      };
    },
    translate: function(rect, x, y) {
      if (x == null) {
        x = 0;
      }
      if (y == null) {
        y = 0;
      }
      return {
        bottom: rect.bottom + y,
        top: rect.top + y,
        left: rect.left + x,
        right: rect.right + x,
        width: rect.width,
        height: rect.height
    };
    },
    subtract: function(rect1, rect2) {
      var rects;
      rect2 = this.create(Math.max(rect1.left, rect2.left), Math.max(rect1.top, rect2.top), Math.min(rect1.right, rect2.right), Math.min(rect1.bottom, rect2.bottom));
      if (rect2.width < 0 || rect2.height < 0) {
        return [Rect.copy(rect1)];
      }
      rects = [this.create(rect1.left, rect1.top, rect2.left, rect2.top), this.create(rect2.left, rect1.top, rect2.right, rect2.top), this.create(rect2.right, rect1.top, rect1.right, rect2.top), this.create(rect1.left, rect2.top, rect2.left, rect2.bottom), this.create(rect2.right, rect2.top, rect1.right, rect2.bottom), this.create(rect1.left, rect2.bottom, rect2.left, rect1.bottom), this.create(rect2.left, rect2.bottom, rect2.right, rect1.bottom), this.create(rect2.right, rect2.bottom, rect1.right, rect1.bottom)];
      return rects.filter(function(rect) {
        return rect.height > 0 && rect.width > 0;
      });
    },
    intersects: function(rect1, rect2) {
      return rect1.right > rect2.left && rect1.left < rect2.right && rect1.bottom > rect2.top && rect1.top < rect2.bottom;
    },
    intersectsStrict: function(rect1, rect2) {
      return rect1.right >= rect2.left && rect1.left <= rect2.right && rect1.bottom >= rect2.top && rect1.top <= rect2.bottom;
    },
    equals: function(rect1, rect2) {
      var i, len, property, ref;
      ref = ["top", "bottom", "left", "right", "width", "height"];
      for (i = 0, len = ref.length; i < len; i++) {
        property = ref[i];
        if (rect1[property] !== rect2[property]) {
          return false;
        }
      }
      return true;
    },
    intersect: function(rect1, rect2) {
      return this.create(Math.max(rect1.left, rect2.left), Math.max(rect1.top, rect2.top), Math.min(rect1.right, rect2.right), Math.min(rect1.bottom, rect2.bottom));
    }
  };

var DomUtils = {
    documentReady: function() {
      var callbacks, isReady, onDOMContentLoaded, ref;
      ref = [document.readyState !== "loading", []], isReady = ref[0], callbacks = ref[1];
      if (!isReady) {
        window.addEventListener("DOMContentLoaded", onDOMContentLoaded = forTrusted(function() {
          var callback, i, len;
          window.removeEventListener("DOMContentLoaded", onDOMContentLoaded);
          isReady = true;
          for (i = 0, len = callbacks.length; i < len; i++) {
            callback = callbacks[i];
            callback();
          }
          return callbacks = null;
        }));
      }
      return function(callback) {
        if (isReady) {
          return callback();
        } else {
          return callbacks.push(callback);
        }
      };
    },
    getVisibleClientRect: function(element, testChildren) {
      var child, childClientRect, clientRect, clientRects, computedStyle, i, isInlineZeroHeight, j, len, len1, ref, ref1;
      if (testChildren == null) {
        testChildren = false;
      }
      clientRects = (function() {
        var i, len, ref, results;
        ref = element.getClientRects();
        results = [];
        for (i = 0, len = ref.length; i < len; i++) {
          clientRect = ref[i];
          results.push(Rect.copy(clientRect));
        }
        return results;
      })();
      isInlineZeroHeight = function() {
        var elementComputedStyle, isInlineZeroFontSize;
        elementComputedStyle = window.getComputedStyle(element, null);
        isInlineZeroFontSize = (0 === elementComputedStyle.getPropertyValue("display").indexOf("inline")) && (elementComputedStyle.getPropertyValue("font-size") === "0px");
        isInlineZeroHeight = function() {
          return isInlineZeroFontSize;
        };
        return isInlineZeroFontSize;
      };
      for (i = 0, len = clientRects.length; i < len; i++) {
        clientRect = clientRects[i];
        if ((clientRect.width === 0 || clientRect.height === 0) && testChildren) {
          ref = element.children;
          for (j = 0, len1 = ref.length; j < len1; j++) {
            child = ref[j];
            computedStyle = window.getComputedStyle(child, null);
            if (computedStyle.getPropertyValue("float") === "none" && !((ref1 = computedStyle.getPropertyValue("position")) === "absolute" || ref1 === "fixed") && !(clientRect.height === 0 && isInlineZeroHeight() && 0 === computedStyle.getPropertyValue("display").indexOf("inline"))) {
              continue;
            }
            childClientRect = this.getVisibleClientRect(child, true);
            if (childClientRect === null || childClientRect.width < 3 || childClientRect.height < 3) {
              continue;
            }
            return childClientRect;
          }
        } else {
          clientRect = this.cropRectToVisible(clientRect);
          if (clientRect === null || clientRect.width < 3 || clientRect.height < 3) {
            continue;
          }
          computedStyle = window.getComputedStyle(element, null);
          if (computedStyle.getPropertyValue('visibility') !== 'visible') {
            continue;
          }
          return clientRect;
        }
      }
      return null;
    },
    cropRectToVisible: function(rect) {
      var boundedRect;
      boundedRect = Rect.create(Math.max(rect.left, 0), Math.max(rect.top, 0), rect.right, rect.bottom);
      if (boundedRect.top >= window.innerHeight - 4 || boundedRect.left >= window.innerWidth - 4) {
        return null;
      } else {
        return boundedRect;
      }
    },
    getClientRectsForAreas: function(imgClientRect, areas) {
      var area, coords, diff, i, len, r, rect, rects, ref, shape, x, x1, x2, y, y1, y2;
      rects = [];
      for (i = 0, len = areas.length; i < len; i++) {
        area = areas[i];
        coords = area.coords.split(",").map(function(coord) {
          return parseInt(coord, 10);
        });
        shape = area.shape.toLowerCase();
        if (shape === "rect" || shape === "rectangle") {
          x1 = coords[0], y1 = coords[1], x2 = coords[2], y2 = coords[3];
        } else if (shape === "circle" || shape === "circ") {
          x = coords[0], y = coords[1], r = coords[2];
          diff = r / Math.sqrt(2);
          x1 = x - diff;
          x2 = x + diff;
          y1 = y - diff;
          y2 = y + diff;
        } else if (shape === "default") {
          ref = [0, 0, imgClientRect.width, imgClientRect.height], x1 = ref[0], y1 = ref[1], x2 = ref[2], y2 = ref[3];
        } else {
          x1 = coords[0], y1 = coords[1], x2 = coords[2], y2 = coords[3];
        }
        rect = Rect.translate(Rect.create(x1, y1, x2, y2), imgClientRect.left, imgClientRect.top);
        rect = this.cropRectToVisible(rect);
        if (rect && !isNaN(rect.top)) {
          rects.push({
            element: area,
            rect: rect
          });
        }
      }
      return rects;
    },
    isSelectable: function(element) {
      var unselectableTypes;
      if (!(element instanceof Element)) {
        return false;
      }
      unselectableTypes = ["button", "checkbox", "color", "file", "hidden", "image", "radio", "reset", "submit"];
      return (element.nodeName.toLowerCase() === "input" && unselectableTypes.indexOf(element.type) === -1) || element.nodeName.toLowerCase() === "textarea" || element.isContentEditable;
    },
    getViewportTopLeft: function() {
      var box, clientLeft, clientTop, marginLeft, marginTop, rect, style;
      box = document.documentElement;
      style = getComputedStyle(box);
      rect = box.getBoundingClientRect();
      if (style.position === "static" && !/content|paint|strict/.test(style.contain || "")) {
        marginTop = parseInt(style.marginTop);
        marginLeft = parseInt(style.marginLeft);
        return {
          top: -rect.top + marginTop,
          left: -rect.left + marginLeft
        };
      } else {
        //if (Utils.isFirefox()) 
        if (true) {
          clientTop = parseInt(style.borderTopWidth);
          clientLeft = parseInt(style.borderLeftWidth);
        } else {
          clientTop = box.clientTop, clientLeft = box.clientLeft;
        }
        return {
          top: -rect.top - clientTop,
          left: -rect.left - clientLeft
        };
      }
    },
    makeXPath: function(elementArray) {
			var element, i, len, xpath;
			xpath = [];
			for (i = 0, len = elementArray.length; i < len; i++) {
							element = elementArray[i];
							xpath.push(".//" + element, ".//xhtml:" + element);
						}
			return xpath.join(" | ");
		},
	evaluateXPath: function(xpath, resultType) {
				var contextNode, namespaceResolver;
				contextNode = document.webkitIsFullScreen ? document.webkitFullscreenElement : document.documentElement;
				namespaceResolver = function(namespace) {
								if (namespace === "xhtml") {
													return "http://www.w3.org/1999/xhtml";
												} else {
																	return null;
																}
							};
				return document.evaluate(xpath, contextNode, namespaceResolver, resultType, null);
			},
 simulateClick: function(element, modifiers) {
       var defaultActionShouldTrigger, event, eventSequence, i, len, results;
       if (modifiers == null) {
               modifiers = {};
             }
       eventSequence = ["mouseover", "mousedown", "mouseup", "click"];
       results = [];
       for (i = 0, len = eventSequence.length; i < len; i++) {
               event = eventSequence[i];
               defaultActionShouldTrigger = /*Utils.isFirefox() &&*/ Object.keys(modifiers).length === 0 && event === "click" && element.target === "_blank" && element.href && !element.hasAttribute("onclick") && !element.hasAttribute("_vimium-has-onclick-listener") ? true : this.simulateMouseEvent(event, element, modifiers);
               if (event === "click" && defaultActionShouldTrigger /*&& Utils.isFirefox()*/) {
                         if (0 < Object.keys(modifiers).length || element.target === "_blank") {
                                     DomUtils.simulateClickDefaultAction(element, modifiers);
                                   }
                       }
               results.push(defaultActionShouldTrigger);
             }
       return results;
     },
      simulateMouseEvent: (function() {
            var lastHoveredElement;
            lastHoveredElement = void 0;
            return function(event, element, modifiers) {
                    var mouseEvent;
                    if (modifiers == null) {
                              modifiers = {};
                            }
                    if (event === "mouseout") {
                              if (element == null) {
                                          element = lastHoveredElement;
                                        }
                              lastHoveredElement = void 0;
                              if (element == null) {
                                          return;
                                        }
                            } else if (event === "mouseover") {
                                      this.simulateMouseEvent("mouseout", void 0, modifiers);
                                      lastHoveredElement = element;
                                    }
                    mouseEvent = document.createEvent("MouseEvents");
                    mouseEvent.initMouseEvent(event, true, true, window, 1, 0, 0, 0, 0, modifiers.ctrlKey, modifiers.altKey, modifiers.shiftKey, modifiers.metaKey, 0, null);
                    return element.dispatchEvent(mouseEvent);
                  };
          })(),
      simulateClickDefaultAction: function(element, modifiers) {
            var altKey, ctrlKey, metaKey, newTabModifier, ref, shiftKey;
            if (modifiers == null) {
                    modifiers = {};
                  }
            if (!(((ref = element.tagName) != null ? ref.toLowerCase() : void 0) === "a" && (element.href != null))) {
                    return;
                  }
            ctrlKey = modifiers.ctrlKey, shiftKey = modifiers.shiftKey, metaKey = modifiers.metaKey, altKey = modifiers.altKey;
            if (KeyboardUtils.platform === "Mac") {
                    newTabModifier = metaKey === true && ctrlKey === false;
                  } else {
                          newTabModifier = metaKey === false && ctrlKey === true;
                        }
            if (newTabModifier) {
                    chrome.runtime.sendMessage({
                              handler: "openUrlInNewTab",
                              url: element.href,
                              active: shiftKey === true
                            });
                  } else if (shiftKey === true && metaKey === false && ctrlKey === false && altKey === false) {
                          chrome.runtime.sendMessage({
                                    handler: "openUrlInNewWindow",
                                    url: element.href
                                  });
                        } else if (element.target === "_blank") {
                                chrome.runtime.sendMessage({
                                          handler: "openUrlInNewTab",
                                          url: element.href,
                                          active: true
                                        });
                              }
          }
}

  var LocalHints = {
    getVisibleClickable: function(element) {
      var actionName, areas, areasAndRects, base1, clientRect, contentEditable, eventType, i, imgClientRects, isClickable, jsactionRule, jsactionRules, len, map, mapName, namespace, onlyHasTabIndex, possibleFalsePositive, reason, ref, ref1, ref10, ref11, ref12, ref2, ref3, ref4, ref5, ref6, ref7, ref8, ref9, role, ruleSplit, tabIndex, tabIndexValue, tagName, visibleElements, slice;
      tagName = (ref = typeof (base1 = element.tagName).toLowerCase === "function" ? base1.toLowerCase() : void 0) != null ? ref : "";
      isClickable = false;
      onlyHasTabIndex = false;
      possibleFalsePositive = false;
      visibleElements = [];
      reason = null;
      slice = [].slice;
      if (tagName === "img") {
        mapName = element.getAttribute("usemap");
        if (mapName) {
          imgClientRects = element.getClientRects();
          mapName = mapName.replace(/^#/, "").replace("\"", "\\\"");
          map = document.querySelector("map[name=\"" + mapName + "\"]");
          if (map && imgClientRects.length > 0) {
            areas = map.getElementsByTagName("area");
            areasAndRects = DomUtils.getClientRectsForAreas(imgClientRects[0], areas);
            visibleElements.push.apply(visibleElements, areasAndRects);
          }
        }
      }
      if (((ref1 = (ref2 = element.getAttribute("aria-hidden")) != null ? ref2.toLowerCase() : void 0) === "" || ref1 === "true") || ((ref3 = (ref4 = element.getAttribute("aria-disabled")) != null ? ref4.toLowerCase() : void 0) === "" || ref3 === "true")) {
        return [];
      }
      if (this.checkForAngularJs == null) {
        this.checkForAngularJs = (function() {
          var angularElements, i, k, len, len1, ngAttributes, prefix, ref5, ref6, separator;
          angularElements = document.getElementsByClassName("ng-scope");
          if (angularElements.length === 0) {
            return function() {
              return false;
            };
          } else {
            ngAttributes = [];
            ref5 = ['', 'data-', 'x-'];
            for (i = 0, len = ref5.length; i < len; i++) {
              prefix = ref5[i];
              ref6 = ['-', ':', '_'];
              for (k = 0, len1 = ref6.length; k < len1; k++) {
                separator = ref6[k];
                ngAttributes.push(prefix + "ng" + separator + "click");
              }
            }
            return function(element) {
              var attribute, l, len2;
              for (l = 0, len2 = ngAttributes.length; l < len2; l++) {
                attribute = ngAttributes[l];
                if (element.hasAttribute(attribute)) {
                  return true;
                }
              }
              return false;
            };
          }
        })();
      }
      isClickable || (isClickable = this.checkForAngularJs(element));
      if (element.hasAttribute("onclick") || (role = element.getAttribute("role")) && ((ref5 = role.toLowerCase()) === "button" || ref5 === "tab" || ref5 === "link" || ref5 === "checkbox" || ref5 === "menuitem" || ref5 === "menuitemcheckbox" || ref5 === "menuitemradio") || (contentEditable = element.getAttribute("contentEditable")) && ((ref6 = contentEditable.toLowerCase()) === "" || ref6 === "contenteditable" || ref6 === "true")) {
        isClickable = true;
      }
      if (!isClickable && element.hasAttribute("jsaction")) {
        jsactionRules = element.getAttribute("jsaction").split(";");
        for (i = 0, len = jsactionRules.length; i < len; i++) {
          jsactionRule = jsactionRules[i];
          ruleSplit = jsactionRule.trim().split(":");
          if ((1 <= (ref7 = ruleSplit.length) && ref7 <= 2)) {
            ref8 = ruleSplit.length === 1 ? ["click"].concat(slice.call(ruleSplit[0].trim().split(".")), ["_"]) : [ruleSplit[0]].concat(slice.call(ruleSplit[1].trim().split(".")), ["_"]), eventType = ref8[0], namespace = ref8[1], actionName = ref8[2];
            isClickable || (isClickable = eventType === "click" && namespace !== "none" && actionName !== "_");
          }
        }
      }
      switch (tagName) {
        case "a":
          isClickable = true;
          break;
        case "textarea":
          isClickable || (isClickable = !element.disabled && !element.readOnly);
          break;
        case "input":
          isClickable || (isClickable = !(((ref9 = element.getAttribute("type")) != null ? ref9.toLowerCase() : void 0) === "hidden" || element.disabled || (element.readOnly && DomUtils.isSelectable(element))));
          break;
        case "button":
        case "select":
          isClickable || (isClickable = !element.disabled);
          break;
        case "label":
          isClickable || (isClickable = (element.control != null) && !element.control.disabled && (this.getVisibleClickable(element.control)).length === 0);
          break;
        case "body":
          isClickable || (isClickable = element === document.body && !windowIsFocused() && window.innerWidth > 3 && window.innerHeight > 3 && ((ref10 = document.body) != null ? ref10.tagName.toLowerCase() : void 0) !== "frameset" ? reason = "Frame." : void 0);
          //isClickable || (isClickable = element === document.body && windowIsFocused() && Scroller.isScrollableElement(element) ? reason = "Scroll." : void 0);
          break;
        case "img":
          isClickable || (isClickable = (ref11 = element.style.cursor) === "zoom-in" || ref11 === "zoom-out");
          break;
        case "div":
        case "ol":
        case "ul":
          //isClickable || (isClickable = element.clientHeight < element.scrollHeight && Scroller.isScrollableElement(element) ? reason = "Scroll." : void 0);
          break;
        case "details":
          isClickable = true;
          reason = "Open.";
      }
      if (!isClickable && 0 <= ((ref12 = element.getAttribute("class")) != null ? ref12.toLowerCase().indexOf("button") : void 0)) {
        possibleFalsePositive = isClickable = true;
      }
      tabIndexValue = element.getAttribute("tabindex");
      tabIndex = tabIndexValue === "" ? 0 : parseInt(tabIndexValue);
      if (!(isClickable || isNaN(tabIndex) || tabIndex < 0)) {
        isClickable = onlyHasTabIndex = true;
      }
      if (isClickable) {
        clientRect = DomUtils.getVisibleClientRect(element, true);
        if (clientRect !== null) {
          visibleElements.push({
            element: element,
            rect: clientRect,
            secondClassCitizen: onlyHasTabIndex,
            possibleFalsePositive: possibleFalsePositive,
            reason: reason
          });
        }
      }
      return visibleElements;
    },
    getLocalHints: function(requireHref) {
      var descendantsToCheck, element, elements, hint, i, k, l, left, len, len1, len2, len3, localHints, m, negativeRect, nonOverlappingElements, position, rects, ref, ref1, top, visibleElement, visibleElements;
      if (!document.documentElement) {
        return [];
      }
      elements = document.documentElement.getElementsByTagName("*");
      visibleElements = [];
      for (i = 0, len = elements.length; i < len; i++) {
        element = elements[i];
        if (!(requireHref && !element.href)) {
          visibleElement = this.getVisibleClickable(element);
          if (requireHref && element.href && visibleElement.length >  0) {
            visibleElement[0]["href"] = element.href;
          }
          visibleElements.push.apply(visibleElements, visibleElement);
        }
      }
      visibleElements = visibleElements.reverse();
      descendantsToCheck = [1, 2, 3];
      visibleElements = (function() {
        var k, len1, results;
        results = [];
        for (position = k = 0, len1 = visibleElements.length; k < len1; position = ++k) {
          element = visibleElements[position];
          if (element.possibleFalsePositive && (function() {
            var _, candidateDescendant, index, l, len2;
            index = Math.max(0, position - 6);
            while (index < position) {
              candidateDescendant = visibleElements[index].element;
              for (l = 0, len2 = descendantsToCheck.length; l < len2; l++) {
                _ = descendantsToCheck[l];
                candidateDescendant = candidateDescendant != null ? candidateDescendant.parentElement : void 0;
                if (candidateDescendant === element.element) {
                  return true;
                }
              }
              index += 1;
            }
            return false;
          })()) {
            continue;
          }
          results.push(element);
        }
        return results;
      })();
      localHints = nonOverlappingElements = [];
      while (visibleElement = visibleElements.pop()) {
        rects = [visibleElement.rect];
        for (k = 0, len1 = visibleElements.length; k < len1; k++) {
          negativeRect = visibleElements[k].rect;
          rects = (ref = []).concat.apply(ref, rects.map(function(rect) {
            return Rect.subtract(rect, negativeRect);
          }));
        }
        if (rects.length > 0) {
					try {
					var testextend = extend(visibleElement, { rect: rects[0] });
					nonOverlappingElements.push(testextend);
					} catch(error) {
						nonOverlappingElements.push(visibleElement);
					}
          /*nonOverlappingElements.push(extend(visibleElement, {
            rect: rects[0]
          }));*/
        } else {
          if (!visibleElement.secondClassCitizen) {
            nonOverlappingElements.push(visibleElement);
          }
        }
      } 
      ref1 = DomUtils.getViewportTopLeft(), top = ref1.top, left = ref1.left;
      for (l = 0, len2 = nonOverlappingElements.length; l < len2; l++) {
        hint = nonOverlappingElements[l];
        hint.rect.top += top;
        hint.rect.left += left;
      }
      /*if (Settings.get("filterLinkHints")) {
        for (m = 0, len3 = localHints.length; m < len3; m++) {
          hint = localHints[m];
          extend(hint, this.generateLinkText(hint));
        }
      }*/
      return localHints;
    },
    generateLinkText: function(hint) {
      var element, linkText, nodeName, ref, showLinkText;
      element = hint.element;
      linkText = "";
      showLinkText = false;
      nodeName = element.nodeName.toLowerCase();
      if (nodeName === "input") {
        if ((element.labels != null) && element.labels.length > 0) {
          linkText = element.labels[0].textContent.trim();
          if (linkText[linkText.length - 1] === ":") {
            linkText = linkText.slice(0, linkText.length - 1);
          }
          showLinkText = true;
        } else if (((ref = element.getAttribute("type")) != null ? ref.toLowerCase() : void 0) === "file") {
          linkText = "Choose File";
        } else if (element.type !== "password") {
          linkText = element.value;
          if (!linkText && 'placeholder' in element) {
            linkText = element.placeholder;
          }
        }
      } else if (nodeName === "a" && !element.textContent.trim() && element.firstElementChild && element.firstElementChild.nodeName.toLowerCase() === "img") {
        linkText = element.firstElementChild.alt || element.firstElementChild.title;
        if (linkText) {
          showLinkText = true;
        }
      } else if (hint.reason != null) {
        linkText = hint.reason;
        showLinkText = true;
      } else if (0 < element.textContent.length) {
        linkText = element.textContent.slice(0, 256);
      } else if (element.hasAttribute("title")) {
        linkText = element.getAttribute("title");
      } else {
        linkText = element.innerHTML.slice(0, 256);
      }
      return {
        linkText: linkText.trim(),
        showLinkText: showLinkText
      };
    }
  };

var VimiumNormal = {
  followLink : function(linkElement) {
      if (linkElement.nodeName.toLowerCase() === "link") {
            return window.location.href = linkElement.href;
          } else {
                linkElement.scrollIntoView();
                return DomUtils.simulateClick(linkElement);
              }
    },
  findAndFollowLink : function(linkStrings) {
      var boundingClientRect, candidateLink, candidateLinks, computedStyle, exactWordRegex, i, j, k, l, len, len1, len2, len3, link, linkMatches, linkString, links, linksXPath, m, n, ref, ref1;
      linksXPath = DomUtils.makeXPath(["a", "*[@onclick or @role='link' or contains(@class, 'button')]"]);
      links = DomUtils.evaluateXPath(linksXPath, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE);
      candidateLinks = [];
      for (i = j = ref = links.snapshotLength - 1; j >= 0; i = j += -1) {
            link = links.snapshotItem(i);
            boundingClientRect = link.getBoundingClientRect();
            if (boundingClientRect.width === 0 || boundingClientRect.height === 0) {
                    continue;
                  }
            computedStyle = window.getComputedStyle(link, null);
            if (computedStyle.getPropertyValue("visibility") !== "visible" || computedStyle.getPropertyValue("display") === "none") {
                    continue;
                  }
            linkMatches = false;
            for (k = 0, len = linkStrings.length; k < len; k++) {
                    linkString = linkStrings[k];
                    if (link.innerText.toLowerCase().indexOf(linkString) !== -1 || 0 <= ((ref1 = link.value) != null ? typeof ref1.indexOf === "function" ? ref1.indexOf(linkString) : void 0 : void 0)) {
                              linkMatches = true;
                              break;
                            }
                  }
            if (!linkMatches) {
                    continue;
                  }
            candidateLinks.push(link);
          }
      if (candidateLinks.length === 0) {
            return;
          }
      for (l = 0, len1 = candidateLinks.length; l < len1; l++) {
            link = candidateLinks[l];
            link.wordCount = link.innerText.trim().split(/\s+/).length;
          }
      candidateLinks.forEach(function(a, i) {
            return a.originalIndex = i;
          });
      candidateLinks = candidateLinks.sort(function(a, b) {
            if (a.wordCount === b.wordCount) {
                    return a.originalIndex - b.originalIndex;
                  } else {
                          return a.wordCount - b.wordCount;
                        }
          }).filter(function(a) {
                return a.wordCount <= candidateLinks[0].wordCount + 1;
              });
      for (m = 0, len2 = linkStrings.length; m < len2; m++) {
            linkString = linkStrings[m];
            exactWordRegex = /\b/.test(linkString[0]) || /\b/.test(linkString[linkString.length - 1]) ? new RegExp("\\b" + linkString + "\\b", "i") : new RegExp(linkString, "i");
            for (n = 0, len3 = candidateLinks.length; n < len3; n++) {
                    candidateLink = candidateLinks[n];
                    if (exactWordRegex.test(candidateLink.innerText) || (candidateLink.value && exactWordRegex.test(candidateLink.value))) {
                              this.followLink(candidateLink);
                              return true;
                            }
                  }
          }
      return false;
    },
  findAndFollowRel : function(value) {
      var element, elements, j, k, len, len1, relTags, tag;
      relTags = ["link", "a", "area"];
      for (j = 0, len = relTags.length; j < len; j++) {
            tag = relTags[j];
            elements = document.getElementsByTagName(tag);
            for (k = 0, len1 = elements.length; k < len1; k++) {
                    element = elements[k];
                    if (element.hasAttribute("rel") && element.rel.toLowerCase() === value) {
                              this.followLink(element);
                              return true;
                            }
                  }
          }
  },
    textInputXPath : function() {
        var inputElements, textInputTypes;
        textInputTypes = ["text", "search", "email", "url", "number", "password", "date", "tel"];
        inputElements = [
              "input[" + "(" + textInputTypes.map(function(type) {
                      return '@type="' + type + '"';
                    }).join(" or ") + "or not(@type))" + " and not(@disabled or @readonly)]", "textarea", "*[@contenteditable='' or translate(@contenteditable, 'TRUE', 'true')='true']"
            ];
        return typeof DomUtils !== "undefined" && DomUtils !== null ? DomUtils.makeXPath(inputElements) : void 0;
      },
  focusInput: function(count) {
    var element, elements, hint, hints, i, recentlyFocusedElement, resultSet, selectedInputIndex, tuple, visibleInputs;
    resultSet = DomUtils.evaluateXPath(textInputXPath, XPathResult.ORDERED_NODE_SNAPSHOT_TYPE);
    visibleInputs = (function() {
      var j, ref, results;
      results = [];
      for (i = j = 0, ref = resultSet.snapshotLength; j < ref; i = j += 1) {
        element = resultSet.snapshotItem(i);
        if (!DomUtils.getVisibleClientRect(element, true)) {
          continue;
        }
        results.push({
          element: element,
          index: i,
          rect: Rect.copy(element.getBoundingClientRect())
        });
      }
      return results;
    })();
    visibleInputs.sort(function(arg, arg1) {
      var element1, element2, i1, i2, tabDifference;
      element1 = arg.element, i1 = arg.index;
      element2 = arg1.element, i2 = arg1.index;
      if (element1.tabIndex > 0) {
        if (element2.tabIndex > 0) {
          tabDifference = element1.tabIndex - element2.tabIndex;
          if (tabDifference !== 0) {
            return tabDifference;
          } else {
            return i1 - i2;
          }
        } else {
          return -1;
        }
      } else if (element2.tabIndex > 0) {
        return 1;
      } else {
        return i1 - i2;
      }
    });
    if (visibleInputs.length === 0) {
      HUD.showForDuration("There are no inputs to focus.", 1000);
      return;
    }
    recentlyFocusedElement = lastFocusedInput();
    selectedInputIndex = count === 1 ? (elements = visibleInputs.map(function(visibleInput) {
      return visibleInput.element;
    }), Math.max(0, elements.indexOf(recentlyFocusedElement))) : Math.min(count, visibleInputs.length) - 1;
    hints = (function() {
      var j, len, results;
      results = [];
      for (j = 0, len = visibleInputs.length; j < len; j++) {
        tuple = visibleInputs[j];
        hint = DomUtils.createElement("div");
        hint.className = "vimiumReset internalVimiumInputHint vimiumInputHint";
        hint.style.left = (tuple.rect.left - 1) + window.scrollX + "px";
        hint.style.top = (tuple.rect.top - 1) + window.scrollY + "px";
        hint.style.width = tuple.rect.width + "px";
        hint.style.height = tuple.rect.height + "px";
        results.push(hint);
      }
      return results;
    })();
    return new FocusSelector(hints, visibleInputs, selectedInputIndex);
  }
  };


export function MiscVimium() {
  if (window.forTrusted == null) {
    window.forTrusted = function(handler) {
      return function(event) {
        if (event != null ? event.isTrusted : void 0) {
          return handler.apply(this, arguments);
        } else {
          return true;
        }
      };
    };
  }
  window.windowIsFocused = function() {
    var windowHasFocus;
    windowHasFocus = null;
    DomUtils.documentReady(function() {
      return windowHasFocus = document.hasFocus();
    });
    window.addEventListener("focus", forTrusted(function(event) {
      if (event.target === window) {
        windowHasFocus = true;
      }
      return true;
    }));
    window.addEventListener("blur", forTrusted(function(event) {
      if (event.target === window) {
        windowHasFocus = false;
      }
      return true;
    }));
    return function() {
      return windowHasFocus;
    };
  };
};

export { Rect }
export { DomUtils }
export { LocalHints }
export { VimiumNormal }


