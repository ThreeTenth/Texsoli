
function api(api) {
  return "/api" + api
}

(function (window) {

  'use strict';

  // class helper functions from bonzo https://github.com/ded/bonzo

  function classReg(className) {
    return new RegExp("(^|\\s+)" + className + "(\\s+|$)");
  }

  // classList support for class management
  // altho to be fair, the api sucks because it won't accept multiple classes at once
  var hasClass, addClass, removeClass;

  hasClass = function (elem, c) {
    return classReg(c).test(elem.className);
  };
  addClass = function (elem, c) {
    if (!hasClass(elem, c)) {
      elem.className = elem.className + ' ' + c;
    }
  };
  removeClass = function (elem, c) {
    elem.className = elem.className.replace(classReg(c), ' ');
  };

  function toggleClass(elem, c) {
    var fn = hasClass(elem, c) ? removeClass : addClass;
    fn(elem, c);
  }

  var classie = {
    // full names
    hasClass: hasClass,
    addClass: addClass,
    removeClass: removeClass,
    toggleClass: toggleClass,
    // short names
    has: hasClass,
    add: addClass,
    remove: removeClass,
    toggle: toggleClass
  };

  // transport
  if (typeof define === 'function' && define.amd) {
    // AMD
    define(classie);
  } else {
    // browser global
    window.classie = classie;
  }

})(window);

var fragmentCache = {}
var fragmentsCache = {}

var app = new Vue({
  el: '#app',
  data: {
    fragment: { id: 0 },
    is404: false
  },
})

var frags = new Vue({
  el: "#frags",
  data: {
    fragments: [],
    prevFragID: 0,
    isShow: true
  },
  methods: {
    switchFragment: switchFragment
  }
})

var editor = new Vue({
  el: "#editor",
  data: {
    title: "",
    message: "",
    placeholder: "",
    prevFragID: 0
  }
})

function addFragment() {
  var editorPage = document.getElementById("editorPage")
  classie.add(editorPage, "md-show")
  editor.prevFragID = frags.prevFragID
  if (0 == frags.prevFragID) {
    editor.title = "新的开始"
    editor.placeholder = "写一个什么样的开头呢？？"
  } else {
    editor.title = "新版本"
    editor.placeholder = "写一个什么不一样的版本呢？"
  }
}

function appendFragment() {
  var editorPage = document.getElementById("editorPage")
  classie.add(editorPage, "md-show")
  editor.prevFragID = app.fragmentID
  editor.placeholder = "写一个更有趣的延续吧！"
  editor.title = "新篇章"
}

function cancelAdd() {
  var editorPage = document.getElementById("editorPage")
  classie.remove(editorPage, "md-show")
}

function submitFragment() {
  axios.post(api('/v1/fragment'), {
    prevID: editor.prevFragID,
    text: editor.message
  })
    .then(function (response) {
      Snackbar.show({ text: "提交成功 😀", showAction: false })
      cancelAdd()
      editor.message = ""
    })
    .catch(function (error) {
      Snackbar.show({ text: "提交失败 🙁", showAction: false })
    });
}

function switchFragment(fragment, fragments = null) {
  app.is404 = false
  app.fragment = fragment

  // if (0 != fragment.prevID) {
  //   const state = {}
  //   const title = ''
  //   const url = "/fragment/" + fragment.id

  //   history.pushState(state, title, url)
  // }

  if (undefined != fragments || null != fragments) {
    frags.fragments = fragments
    frags.prevFragID = fragment.prevID

    fragmentsCache[fragment.prevID] = fragments
  }

  fragmentCache[fragment.id] = fragment
}

function showFrags() {
  frags.isShow = !frags.isShow
}

function prev() {
  // 上一节逻辑，如果当前页面是 404，则直接隐藏即可，当前 app.fragment 持有的对象便是上一节片段
  if (app.is404) {
    app.is404 = false
  } else {
    var prevFragment = fragmentCache[app.fragment.prevID]
    switchFragment(prevFragment, fragmentsCache[prevFragment.prevID])
  }
}

function next() {
  var currFragID = app.fragment.id
  var url = api('/v1/fragments')
  url += 0 == currFragID ? "" : "/" + currFragID
  axios.get(url)
    .then(function (response) {
      var fragments = response.data

      switchFragment(fragments[0], fragments)
    })
    .catch(function (error) {
      app.is404 = true
    });
}

next()