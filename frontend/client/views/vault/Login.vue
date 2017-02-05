<template>
  <div>
    <div class="tile is-ancestor">

      <div id="loginapp" class="tile is-parent is-6">
        <article class="tile is-child box">
          <h1 class="title">Vault Login</h1>
          <div class="block">
            <form id="form" v-on:submit.prevent="login">

              <!-- to do: display server vault address -->

              <div class="control">
                <label class="label">Authentication Type</label>
                <div class="select is-fullwidth">
                  <select v-model="type">
                    <option>Token</option>
                  </select>
                </div>
              </div>

              <p class="control has-icon">
                <input class="input" type="password" placeholder="Vault Token" v-model="vaultToken">
                <i class="fa fa-lock"></i>
              </p>
              <p class="control">
                <button type="submit" value="Login" class="button is-success">
                  Login
                </button>
              </p>

            </form>
          </div>
        </article>
      </div>

    </div>
  </div>
</template>

<script>
  import Vue from 'vue'

  export default {
    data () {
      return {
        type: 'Token',
        vaultToken: '',
        statusText: ''
      }
    },
    created: function () {
      // fetch CSRF token and push it as an interceptor for POST request upon login
      this.$http.get('/api/login/csrf').then(function (response) {
        Vue.http.interceptors.push((request, next) => {
          request.headers.set('X-CSRF-Token', response.headers.get('x-csrf-token'))
          next()
        })
      }, function (err) {
        this.statusText = err.statusText
        console.log(err.statusText)
      })
    },
    methods: {
      login: function () {
        this.statusText = ''
        var payload = {
          Type: this.type.toLowerCase(),
          ID: this.vaultToken
        }
        this.$http.post('/api/login', payload).then(function (response) {
          console.log(response.data.status)
          this.$router.push({
            name: 'Users'
          })
          this.$router.go(1)
        }, function (err) {
          this.statusText = err.statusText
          console.log(err.statusText)
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
</style>
