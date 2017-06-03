<template>
  <div>
    <div class="tile is-ancestor is-vertical">
      <div class="tile is-parent">
        <article class="tile is-child box is-vertical">

          <!-- nav bar tile -->
          <div class="tile is-parent">
            <div class="tile is-child box">
              <nav class="level-left">
                <div class="level-item">
                  <p class="subtitle is-5">
                    <strong>Goldfish is configured to use 'userTransit'</strong>
                    <a class="is-danger">
                    <span class="icon" @click="changeKey()">
                      <i class="fa fa-pencil-square-o"></i>
                    </span>
                    </a>
                    <strong>key by default</strong>
                  </p>
                </div>
              </nav>
            </div>
          </div>

          <!-- encrypt & decrypt tiles -->
          <div class="tile">

            <!-- encrypt tile -->
            <article class="tile is-parent is-6">
              <div class="tile is-child box">
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
              </div>
            </article>

            <!-- decrypt tile -->
            <article class="tile is-parent is-6">
              <div class="tile is-child box">
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
              </div>
            </article>

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
    },

    changeKey: function () {
      this.$notify({
        title: 'Under Construction',
        message: 'Changeable transit key will come soon',
        type: 'warning'
      })
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
