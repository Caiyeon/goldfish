# Notification

Notification component for Vue Bulma.


## Installation

```
$ npm install vue-bulma-notification --save
```


## Examples

```vue
<template>
  <div>
    <notification :title="'Normal'" :direction="'Down'" :message="'Lorem ipsum dolor sit amet, consectetur adipiscing elit lorem ipsum dolor sit amet, consectetur adipiscing elit'" :duration="0"></notification>
    <button class="button" @click="openNotificationWithType('')">Normal</button>
    <button class="button is-primary" @click="openNotificationWithType('primary')">Primary</button>
    <button class="button is-info" @click="openNotificationWithType('info')">Info</button>
    <button class="button is-success" @click="openNotificationWithType('success')">Success</button>
    <button class="button is-warning" @click="openNotificationWithType('warning')">Warning</button>
    <button class="button is-danger" @click="openNotificationWithType('danger')">Danger</button>
  </div>
</template>

<script>
import Vue from 'vue'
import Notification from 'vue-bulma-notification'

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

export default {
  components: {
    Notification
  },

  mounted () {
    openNotification({
      message: 'Success lorem ipsum dolor sit amet, consectetur adipiscing elit lorem ipsum dolor sit amet, consectetur adipiscing elit',
      type: 'success',
      duration: 0
    })
  },

  methods: {
    openNotificationWithType (type) {
      openNotification({
        title: 'This is a title',
        message: 'This is the message.',
        type: type
      })
    }
  }

}
</script>
```


## Badges

![](https://img.shields.io/badge/license-MIT-blue.svg)
![](https://img.shields.io/badge/status-stable-green.svg)

---

> [fundon.me](https://fundon.me) &nbsp;&middot;&nbsp;
> GitHub [@fundon](https://github.com/fundon) &nbsp;&middot;&nbsp;
> Twitter [@_fundon](https://twitter.com/_fundon)
