<template>
  <div>
    <div class="tile is-ancestor is-vertical">

      <div class="tile is-parent">
        <div class="tile is-parent is-vertical is-6">
          <article class="tile is-child box">
            <h4 class="title is-4">Encrypt</h4>
            <p class="control">
              <textarea v-model="plaintext" class="textarea" placeholder="Paste something here"></textarea>
            </p>
            <p class="control has-addons has-addons-right">
              <a @click="encryptText" class="button is-primary is-outlined">
                Encrypt
                <span class="icon is-small">
                  <i class="fa fa-check"></i>
                </span>
              </a>
              <a @click="clearPlaintext" class="button is-danger is-outlined">
                Clear
                <span class="icon is-small">
                  <i class="fa fa-times"></i>
                </span>
              </a>
            </p>
          </article>
        </div>

        <div class="tile is-parent is-vertical is-6">
          <article class="tile is-child box">
            <h4 class="title is-4">Decrypt</h4>
            <p class="control">
              <textarea v-model="cipher" class="textarea" placeholder="Paste something here"></textarea>
            </p>
            <p class="control has-addons has-addons-right">
              <a @click="decryptText" class="button is-primary is-outlined">
                Decrypt
                <span class="icon is-small">
                  <i class="fa fa-check"></i>
                </span>
              </a>
              <a @click="clearCipher" class="button is-danger is-outlined">
                Clear
                <span class="icon is-small">
                  <i class="fa fa-times"></i>
                </span>
              </a>
            </p>
          </article>
        </div>
      </div>

    <div class="tile is-parent">
      <div class="tile is-parent is-child">
        <article class="tile is-child box">
          <h4 class="title is-4">This tool uses the transit backend to encrypt and decrypt arbitrary strings</h4>
        </article>
      </div>
    </div>

    </div>
  </div>
</template>

<script>
  import Tooltip from 'vue-bulma-tooltip'
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
      Tooltip
    },

    data () {
      return {
        csrf: '',
        plaintext: '',
        cipher: ''
      }
    },

    computed: {
    },

    filters: {
    },

    mounted: function () {
      this.$http.get('/api/transit').then(function (response) {
        this.csrf = response.headers.get('x-csrf-token')
      }, function (err) {
        openNotification({
          title: 'Error',
          message: err.body.error,
          type: 'danger'
        })
        console.log(err.body.error)
      })
    },

    methods: {
      encryptText: function () {
        var body = {
          'Str': this.plaintext
        }
        var headers = {
          headers: {
            'X-CSRF-Token': this.csrf
          }
        }
        this.$http.post('/api/transit/encrypt', body, headers).then(function (response) {
          this.cipher = response.data.result
          this.plaintext = ''
          openNotification({
            title: 'Success',
            message: 'Encryption successful',
            type: 'success'
          })
        }, function (err) {
          openNotification({
            title: 'Error',
            message: err.body.error,
            type: 'danger'
          })
          console.log(err.body.error)
        })
      },

      decryptText: function () {
        var body = {
          'Str': this.cipher
        }
        var headers = {
          headers: {
            'X-CSRF-Token': this.csrf
          }
        }
        this.$http.post('/api/transit/decrypt', body, headers).then(function (response) {
          this.plaintext = response.data.result
          this.cipher = ''
          openNotification({
            title: 'Success',
            message: 'Decryption successful',
            type: 'success'
          })
        }, function (err) {
          openNotification({
            title: 'Error',
            message: err.body.error,
            type: 'danger'
          })
          console.log(err.body.error)
        })
      },

      clearPlaintext: function () {
        this.plaintext = ''
      },
      clearCipher: function () {
        this.cipher = ''
      }
    }
  }
</script>

<style scoped>
  .button {
    margin: 5px 0 0;
  }

  .control .button {
    margin: inherit;
  }

  .fa-trash-o {
    color: red;
  }

  .fa-info {
    color: lightskyblue;
  }
</style>
