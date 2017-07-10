<template>
  <section class="hero is-bold app-navbar animated" :class="{ slideInDown: show, slideOutDown: !show }">
    <div class="hero-head">
      <nav class="navbar">

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
                  <div>Token expires in<p class="has-text-info">{{tokenExpiresIn || 'never'}}</p></div>
                </div>
              </div>
            </div>
          </div>

          <!-- rightside -->
          <div class="navbar-end">
            <div class="navbar-item has-dropdown is-hoverable">
              <a class="navbar-link is-active">
                Docs
              </a>
              <div class="navbar-dropdown ">
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
      now: moment()
    }
  },

  mounted: function () {
    // refresh current time every second, since time is not reactive
    setInterval(() => {
      this.now = moment()
    }, 1000)

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
    }
  },

  methods: {
    ...mapActions([
      'toggleSidebar'
    ])
  }
}
</script>

<style lang="scss">
@import '~bulma/sass/utilities/variables';

.app-navbar {
  position: fixed;
  min-width: 100%;
  z-index: 1024;
  box-shadow: 0 2px 3px rgba(17, 17, 17, 0.1), 0 0 0 1px rgba(17, 17, 17, 0.1);

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
