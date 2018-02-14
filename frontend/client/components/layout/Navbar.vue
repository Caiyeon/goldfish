<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">
      <nav class="navbar is-transparent">

        <div class="navbar-brand">
          <a class="navbar-item" href="/">
            <img src="~assets/logo.svg" :alt="pkginfo.description">
            &nbsp;<span style="color:hsl(171, 100%, 41%)">Goldfish</span>
          </a>

          <a class="navbar-item is-hidden-desktop"
          href="https://github.com/Caiyeon/goldfish" target="_blank">
            <span class="icon" style="color: #333;">
              <i class="fa fa-github"></i>
            </span>
          </a>

          <div class="navbar-burger burger"
          @click="toggleSidebar(!sidebar.opened)"
          data-target="navMenuExample">
            <span></span>
            <span></span>
            <span></span>
          </div>
        </div>

        <div class="navbar-menu">

          <!-- leftside -->
          <div class="navbar-start">
            <!-- session dropdown -->
            <div v-if="session" class="navbar-item has-dropdown is-hoverable">
              <a class="navbar-link is-active">
                Session
              </a>
              <div class="navbar-dropdown is-boxed">
                <div class="navbar-item">
                  <div>{{session.display_name}}
                    <p v-if="tokenExpiresIn === ''" class="has-text-info">will not expire</p>
                    <p v-if="tokenExpiresIn !== ''" class="has-text-info">expires {{tokenExpiresIn}}</p>
                  </div>
                </div>

                <hr v-if="session !== null" class="navbar-divider">
                <div v-if="session !== null" class="navbar-item">
                  <div class="navbar-content">
                    <div class="level">

                      <div class="level-left">
                        <div class="level-item">
                          <button class="button is-primary is-small"
                            @click="renewLogin()" :disabled="!session.renewable">
                            Renew
                          </button>
                        </div>
                      </div>&nbsp;&nbsp;&nbsp;

                      <div class="level-right">
                        <div class="level-item">
                          <p v-if="session !== null" class="control">
                            <button class="button is-warning is-small"
                             @click="logout()">
                              Logout
                            </button>
                          </p>
                        </div>
                      </div>

                    </div>
                  </div>
                </div>

              </div> <!-- end navbar dropdown -->

            </div>
          </div>

          <!-- rightside -->
          <div class="navbar-end">
            <div class="navbar-item" v-if="holidaySeasons">
              <div class="tags has-addons">
                <span class="tag is-success"><i class="fa fa-gift"></i>&nbsp;&nbsp;Happy Holidays!</span>
                <span class="tag is-danger">Happy New Years!&nbsp;&nbsp;<i class="fa fa-smile-o"></i></span>
              </div>
            </div>

            <div class="navbar-item" v-if="updateAvailable">
              <div class="tags has-addons">
                <a class="tag is-primary" :href="latestRelease['html_url']" target="_blank">Update Available</a>
                <a class="tag is-info" :href="latestRelease['html_url']" target="_blank">{{latestRelease.tag_name}}</a>
              </div>
            </div>

            <div class="navbar-item has-dropdown is-hoverable">
              <a class="navbar-link is-active">
                Docs
              </a>
              <div class="navbar-dropdown is-boxed">
                <a class="navbar-item" target="_blank" href="https://github.com/Caiyeon/goldfish/wiki/Configuration#run-time-configurations">
                  Configuration
                </a>
                <a class="navbar-item" target="_blank" href="https://github.com/Caiyeon/goldfish/wiki/Features">
                  Features
                </a>
                <a class="navbar-item" target="_blank" href="https://github.com/Caiyeon/goldfish">
                  Source
                </a>
                <hr class="navbar-divider">
                <div class="navbar-item">
                  <div>Version <p class="has-text-info">{{pkginfo.version}}</p></div>
                </div>
              </div>
            </div>

            <!-- github button -->
            <div class="navbar-item">
              <div class="field is-grouped">
                <p class="control">
                  <a class="button is-info is-outlined"
                  href="https://github.com/Caiyeon/goldfish"
                  target="_blank">
                    <span class="icon">
                      <i class="fa fa-github"></i>
                    </span>
                    <span class="is-hidden-mobile">Source Code</span>
                  </a>
                </p>
              </div>
            </div>
          </div>

        </div>

      </nav>
    </div>
  </section>
</template>

<script>
import { mapGetters, mapActions } from 'vuex'
import moment from 'moment'

export default {
  props: {
    show: Boolean
  },

  data () {
    return {
      profileDropdown: false,
      position: ['center', 'bottom', 'center', 'top'],
      now: moment(),
      latestRelease: {},
      expirationWarning: false
    }
  },

  mounted: function () {
    // refresh current time every second, since time is not reactive
    setInterval(() => {
      this.now = moment()
    }, 1000)

    // check every 30s if the user's token will expire in 10 mins
    setInterval(() => {
      if (this.session === null || this.session['token_expiry'] === 'never') {
        return
      }
      // if token will expire in 10 minutes
      let tokenExpires = moment(this.session['token_expiry'], 'ddd, h:mm:ss A MMMM Do YYYY')
      if (tokenExpires.diff(this.now) < 600000) {
        // show user warning
        if (!this.expirationWarning) {
          this.$message({
            message: 'Your token expires in 10 minutes',
            type: 'warning',
            duration: 0,
            showCloseButton: true
          })
          // set warning flag true so that user does not continue to get warnings
          this.expirationWarning = true
        }
      } else {
        // if user's token has been renewed after warning, reset warning flag
        this.expirationWarning = false
      }
    }, 30000)

    // if session cookie is still valid, load session data
    let raw = window.localStorage.getItem('session')
    if (raw) {
      var session = JSON.parse(raw)
      if (moment().isAfter(moment(session['cookie_expiry'], 'ddd, h:mm:ss A MMMM Do YYYY'))) {
        window.localStorage.removeItem('session')
        this.$store.commit('clearSession')
      } else {
        this.$store.commit('setSession', session)
      }
    } else {
      this.$store.commit('clearSession')
    }

    // on load, check for latest stable release
    this.$http.get('https://api.github.com/repos/caiyeon/goldfish/releases/latest')
    .then((response) => {
      this.latestRelease = response.data
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  computed: {
    ...mapGetters({
      session: 'session',
      pkginfo: 'pkg',
      sidebar: 'sidebar'
    }),

    tokenExpiresIn: function () {
      if (this.session === null || this.session['token_expiry'] === 'never') {
        return ''
      }
      return this.now.to(moment(this.session['token_expiry'], 'ddd, h:mm:ss A MMMM Do YYYY'))
    },

    // parses current package info vs latest stable release to detect if an update is available
    updateAvailable: function () {
      if (this.latestRelease && this.latestRelease.tag_name) {
        // split curr version into 3 numbers
        var curr = this.pkginfo.version
        if (curr.substr(0, 1) === 'v') {
          curr = curr.substr(1)
        }
        var currV = curr.split('-')[0].split('.').map(Number)

        // split new version into 3 numbers
        var newest = this.latestRelease.tag_name
        if (newest.substr(0, 1) === 'v') {
          newest = newest.substr(1)
        }
        var newestV = newest.split('-')[0].split('.').map(Number)

        // convert string numbers into ints and compare according to rank
        if (newestV[0] > currV[0]) {
          return true
        } else if (newestV[0] === currV[0]) {
          if (newestV[1] > currV[1]) {
            return true
          } else if (newestV[1] === currV[1] && newestV[2] > currV[2]) {
            return true
          }
        }
      }
      return false
    },

    // computed client-side only :)
    holidaySeasons: function () {
      return moment().isBetween(
        moment('22/12/2017', 'DD/MM/YYYY'),
        moment('3/1/2018', 'DD/MM/YYYY'),
        'days', '[]'
      )
    }
  },

  methods: {
    ...mapActions([
      'toggleSidebar'
    ]),

    logout: function () {
      // purge session from localstorage
      window.localStorage.removeItem('session')
      // mutate vuex state
      this.$store.commit('clearSession')
    },

    renewLogin: function () {
      this.$http.post('/v1/login/renew-self', {}, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        // deep copy session, update fields, and mutate state
        let newSession = JSON.parse(JSON.stringify(this.session))

        newSession['meta'] = response.data.result['meta']
        newSession['policies'] = response.data.result['policies']
        newSession['token_expiry'] = response.data.result['ttl'] === 0 ? 'never' : moment().add(response.data.result['ttl'], 'seconds').format('ddd, h:mm:ss A MMMM Do YYYY')

        window.localStorage.setItem('session', JSON.stringify(newSession))
        this.$store.commit('setSession', newSession)
        this.$notify({
          title: 'Renew success!',
          message: '',
          type: 'success'
        })
      })
      .catch((error) => {
        this.$onError(error)
      })
    }
  }
}
</script>

<style lang="scss">
.app-navbar {
  position: fixed;
  min-width: 100%;
  z-index: 8;
  box-shadow: 0 2px 3px rgba(17, 17, 17, 0.1), 0 0 0 1px rgba(17, 17, 17, 0.1);
  background-color: white;

  .container {
    margin: auto 10px;
  }

  .nav-right {
    align-items: stretch;
    align-items: stretch;
    flex: 1;
    justify-content: flex-end;
    overflow: hidden;
    overflow-x: auto;
    white-space: nowrap;
  }
}
</style>
