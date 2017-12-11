import Vue from 'vue'
import axios from 'axios'
import NProgress from 'vue-nprogress'
import { sync } from 'vuex-router-sync'
import App from './App.vue'
import router from './router'
import store from './store'
import * as filters from './filters'
import { TOGGLE_SIDEBAR } from 'vuex-store/mutation-types'
import Notification from './views/vue_bulma_modules/vue-bulma-notification'
import Message from 'vue-bulma-message'
import hljs from 'highlight.js'

Vue.prototype.$http = axios
Vue.axios = axios
Vue.use(NProgress)

// Enable devtools
Vue.config.devtools = true

sync(store, router)

const nprogress = new NProgress({ parent: '.nprogress-container' })

const { state } = store

router.beforeEach((route, redirect, next) => {
  if (state.app.device.isMobile && state.app.sidebar.opened) {
    store.commit(TOGGLE_SIDEBAR, false)
  }
  next()
})

Object.keys(filters).forEach(key => {
  Vue.filter(key, filters[key])
})

const app = new Vue({
  router,
  store,
  nprogress,
  ...App
})

const NotificationComponent = Vue.extend(Notification)
const openNotification = (propsData = {
  title: '',
  message: '',
  type: '',
  direction: '',
  duration: 4500,
  container: '.notifications'
}) => {
  return new NotificationComponent({
    el: document.createElement('div'),
    propsData
  })
}
Vue.prototype.$notify = openNotification

function handleError (error) {
  // if the server gave a response message, print that
  if (error.response.data.error) {
    // duration should be proportional to the error message length
    openNotification({
      title: 'Error: ' + error.response.status,
      message: error.response.data.error,
      type: 'danger',
      duration: error.response.data.error.length > 60 ? 20000 : 4500
    })
    console.log(error.response.data.error)
  } else {
    if (error.response.status === 404) {
      openNotification({
        title: 'Error: 404',
        message: 'Not found',
        type: 'danger'
      })
    } else if (error.response.status === 504) {
    // Gateway timeout means vault is down/unreachable
      openNotification({
        title: 'Error: 504',
        message: 'Goldfish or vault is unreachable',
        type: 'danger'
      })
    } else {
      openNotification({
        title: 'Error: ' + error.response.status.toString(),
        message: '',
        type: 'danger'
      })
    }
    console.log(error.response.data)
  }
}
Vue.prototype.$onError = handleError

const MessageComponent = Vue.extend(Message)
const openMessage = (propsData = {
  title: '',
  message: '',
  type: '',
  direction: '',
  duration: 1500,
  container: '.messages'
}) => {
  return new MessageComponent({
    el: document.createElement('div'),
    propsData
  })
}
Vue.prototype.$message = openMessage

Vue.directive('highlightjs', {
  deep: true,
  bind: function (el, binding) {
    // on first bind, highlight all targets
    let targets = el.querySelectorAll('code')
    targets.forEach((target) => {
      // if a value is directly assigned to the directive, use this
      // instead of the element content.
      if (binding.value) {
        target.textContent = binding.value
      }
      hljs.highlightBlock(target)
    })
  },
  componentUpdated: function (el, binding) {
    // after an update, re-fill the content and then highlight
    let targets = el.querySelectorAll('code')
    targets.forEach((target) => {
      if (binding.value) {
        target.textContent = binding.value
        hljs.highlightBlock(target)
      }
    })
  }
})

// custom directive to set focus on an element when it is loaded into the DOM
Vue.directive('focus', {
  inserted: function (el) {
    el.focus()
  }
})

export { app, router, store }
