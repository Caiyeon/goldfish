<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">

        <div class="box">
          <p class="title is-3 is-spaced">Encryption as a service</p>
          <p class="subtitle is-5">Logged in user must have write access to transit key: '/{{ userTransitKey }}' </p>
        </div>

        <div class="tile is-parent is-marginless is-paddingless">
          <div class="tile is-parent is-vertical is-6">
            <article class="tile is-child box">

              <h3 class="title is-3">Encrypt</h3>

              <div class="field">
                <p class="control">
                  <textarea v-model="plaintext" class="textarea" placeholder="Paste something here"></textarea>
                </p>
              </div>

              <div class="field is-pulled-right">
                <p class="control">
                  <a @click="encryptText" class="button is-primary is-outlined">
                    <span>Encrypt</span>
                    <span class="icon">
                      <i class="fa fa-check"></i>
                    </span>
                  </a>
                  <a @click="clearPlaintext" class="button is-danger is-outlined">
                    <span>Clear</span>
                    <span class="icon">
                      <i class="fa fa-times"></i>
                    </span>
                  </a>
                </p>
              </div>

            </article>
          </div>

          <div class="tile is-parent is-vertical is-6">
            <article class="tile is-child box">

              <h3 class="title is-3">Decrypt</h3>

              <div class="field">
                <p class="control">
                  <textarea v-model="cipher" class="textarea" placeholder="Paste something here"></textarea>
                </p>
              </div>

              <div class="field is-pulled-right">
                <p class="control">
                  <a @click="decryptText" class="button is-primary is-outlined">
                    <span>Decrypt</span>
                    <span class="icon">
                      <i class="fa fa-check"></i>
                    </span>
                  </a>
                  <a @click="clearCipher" class="button is-danger is-outlined">
                    <span>Clear</span>
                    <span class="icon">
                      <i class="fa fa-times"></i>
                    </span>
                  </a>
                </p>
              </div>
            </article>
          </div>
        </div>

        </article>
      </div>
    </div>
  </div>
</template>

<script>
import Tooltip from 'vue-bulma-tooltip'
const querystring = require('querystring')

export default {
  components: {
    Tooltip
  },

  data () {
    return {
      csrf: '',
      plaintext: '',
      cipher: '',
      userTransitKey: ''
    }
  },

  mounted: function () {
    this.$http.get('/api/transit').then((response) => {
      this.csrf = response.headers['x-csrf-token']
      this.userTransitKey = response.headers['usertransitkey']
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  methods: {
    encryptText: function () {
      this.$http.post('/api/transit/encrypt', querystring.stringify({
        plaintext: this.plaintext
      }), {
        headers: {'X-CSRF-Token': this.csrf}
      })

      .then((response) => {
        this.cipher = response.data.result
        this.plaintext = ''
        this.$notify({
          title: 'Success',
          message: 'Encryption successful',
          type: 'success'
        })
      })

      .catch((error) => {
        this.$onError(error)
      })
    },

    decryptText: function () {
      this.$http.post('/api/transit/decrypt', querystring.stringify({
        cipher: this.cipher
      }), {
        headers: {'X-CSRF-Token': this.csrf}
      })

      .then((response) => {
        this.plaintext = response.data.result
        this.cipher = ''
        this.$notify({
          title: 'Success',
          message: 'Decryption successful',
          type: 'success'
        })
      })

      .catch((error) => {
        this.$onError(error)
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
