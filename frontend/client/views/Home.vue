<template>
  <div class="content has-text-centered">
    <p>
      <img width="200" src="~assets/logo.svg" :alt="description">
    </p>

    <h1 class="is-title is-bold">{{ name.replace('-', ' ') }}</h1>

    <p>
      <strong>{{ description }}</strong>,
      <a :href="homepage">Live Demo</a>
    </p>

    <p>Supports Vue 2.0 and Bulma 0.3!</p>
  </div>
</template>

<script>
  import Vue from 'vue'
  import Message from 'vue-bulma-message'

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

  export default {
    components: {
      Message
    },

    data () {
      return this.$store.state.pkg
    },

    mounted () {
      openMessage({
        message: 'This goldfish demo is for display purposes only. While it is hooked up to a real vault instance, it should not be considered secure. Head to the Login page to authenticate first!',
        type: 'success',
        duration: 0,
        showCloseButton: true
      })
    }
  }
</script>

<style lang="scss" scoped>
.is-title {
  text-transform: capitalize;
}
</style>
