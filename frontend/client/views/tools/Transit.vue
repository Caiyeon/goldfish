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
                    Goldfish is using transit key <strong>{{editing ? ' ' : userTransitKey}}</strong>
                      &nbsp;<p v-if="editing" class="control">
                        <input class="input is-small"
                        type="text" placeholder="Enter transit key"
                        v-model="userTransitKey"
                        @keyup.enter="editing = false">
                      </p>
                    <a v-if="!editing"><span class="icon" @click="changeKey()">
                      <i class="fa fa-pencil-square-o"></i>
                    </span></a>
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
                <h4 class="subtitle is-4">Encrypt</h4>

                <div class="field">
                  <p class="control">
                    <textarea v-model="plaintext"
                    class="textarea"
                    placeholder="Paste something here"
                    rows="10"></textarea>
                  </p>
                </div>

                <div class="field is-pulled-right">
                  <p class="control">
                    <a @click="encryptText"
                    class="button is-primary is-outlined"
                    :disabled="editing">
                      <span>Encrypt</span>
                      <span class="icon">
                        <i class="fa fa-check"></i>
                      </span>
                    </a>
                    <a @click="clearPlaintext"
                    class="button is-danger is-outlined">
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
                <h4 class="subtitle is-4">Decrypt</h4>

                <div class="field">
                  <p class="control">
                    <textarea v-model="cipher"
                    class="textarea"
                    placeholder="Paste something here"
                    rows="10"></textarea>
                  </p>
                </div>

                <div class="field is-pulled-right">
                  <p class="control">
                    <a @click="decryptText"
                    class="button is-primary is-outlined"
                    :disabled="editing">
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
const querystring = require('querystring')

export default {
  data () {
    return {
      plaintext: '',
      cipher: '',
      userTransitKey: '',
      editing: false
    }
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    }
  },

  mounted: function () {
    this.$http.get('/v1/transit', {
      headers: {'X-Vault-Token': this.session ? this.session.token : ''}
    }).then((response) => {
      this.userTransitKey = response.headers['usertransitkey']
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  methods: {
    encryptText: function () {
      if (this.editing) {
        return
      }

      this.$http.post('/v1/transit/encrypt', querystring.stringify({
        plaintext: this.plaintext,
        key: this.userTransitKey
      }), {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
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
      if (this.editing) {
        return
      }

      this.$http.post('/v1/transit/decrypt', querystring.stringify({
        cipher: this.cipher,
        key: this.userTransitKey
      }), {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
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
      this.editing = true
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
