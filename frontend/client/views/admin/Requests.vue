<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <div class="field">
          <!-- <label class="label">Search by change ID:</label> -->
            <div class="field has-addons">
              <p v-if="request === null" class="control has-icons-right">
                <span class="select">
                  <select v-model="searchType">
                    <option v-bind:value="'changeid'">Change ID</option>
                    <option v-bind:value="'commit'">Commit Hash</option>
                  </select>
                </span>
              </p>
              <p v-else class="control">
                <a class="button is-danger" @click="request = null">
                  Reset
                </a>
              </p>
              <p class="control">
                <input class="input" type="text"
                placeholder="Enter a change ID"
                v-model="searchString"
                @keyup.enter="search()"
                :disabled="request !== null">
              </p>
              <p class="control">
                <a class="button is-info" @click="search()" :disabled="request !== null">
                  Search
                </a>
              </p>
            </div>
          </div>

          <!-- Displaying a single request via changeid -->
          <article v-if="request !== null && !request.length">
            <br>
            <article class="message is-primary">
              <div class="message-body">
                <strong>Requester display name: </strong>{{request.Requester}}<br>
                <strong>Requester accessor hash: </strong>{{request.RequesterHash}}<br>
                <strong>Policy: </strong>{{request.Policy}}<br>
                <strong>Unseal progress: </strong>
                {{progress}} out of {{required}} <strong>{{progress === required ? ' Done!' : ''}}</strong>
              </div>
            </article>

            <div class="field is-grouped">
              <p class="control">
                <button class="button is-success" @click="bConfirm = true">Approve</button>
              </p>
              <p class="control">
                <button v-if="!bReject" class="button is-warning" @click="bReject = true">Reject</button>
                <button v-else class="button is-danger" @click="reject()">Confirm Reject</button>
              </p>
              <div v-if="bConfirm" class="field has-addons">
                <p class="control">
                  <input class="input" type="text"
                  placeholder="Enter an unseal token"
                  v-model="unsealToken"
                  @keyup.enter="approve()">
                </p>
                <p class="control">
                  <a class="button is-info" @click="approve()">
                    Confirm
                  </a>
                </p>
              </div>
            </div>

            <div class="columns">
              <div class="column">
                <article class="message is-primary">
                  <div class="message-header">
                    Current policy rules
                  </div>
                  <pre v-highlightjs="request.Current"><code class="javascript"></code></pre>
                </article>
              </div>

              <div class="column">
                <article class="message is-info">
                  <div class="message-header">
                  Proposed policy rules
                  </div>
                  <pre v-highlightjs="request.New"><code class="javascript"></code></pre>
                </article>
              </div>
            </div>
          </article>

          <!-- Displaying a set of changes via github commit hash -->
          <article v-if="request !== null && request.length > 1">
            <br>
            <article class="message is-primary">
              <div class="message-body">
                <strong>Github commit hash: </strong>{{searchString}}<br>
                <strong>Number of policies affected: </strong>{{request.length}}<br>
                <strong>Unseal progress: </strong>
                {{progress}} out of {{required}} <strong>{{progress === required ? ' Done!' : ''}}</strong>
              </div>
            </article>

            <div class="field is-grouped">
              <p class="control">
                <button class="button is-success" @click="bConfirm = true">Approve</button>
              </p>
              <div v-if="bConfirm" class="field has-addons">
                <p class="control">
                  <input class="input" type="text"
                  placeholder="Enter an unseal token"
                  v-model="unsealToken"
                  @keyup.enter="approve()">
                </p>
                <p class="control">
                  <a class="button is-info" @click="approve()">
                    Confirm
                  </a>
                </p>
              </div>
            </div>

            <div v-for="(policy, index) in request" class="box">
              <!-- policy name title and status tag -->
              <nav class="level">
                <div class="level-left">
                  <div class="level-item">
                    <div class="content is-marginless is-paddingless">
                      <h3 class="is-marginless is-paddingless">{{policy.Policy}}</h3>
                    </div>
                  </div>
                  <div class="level-item">
                    <span v-if="policy.New && policy.Current" class="tag is-info">Will be changed!</span>
                    <span v-if="!policy.New && policy.Current" class="tag is-danger">Will be deleted!</span>
                    <span v-if="policy.New && !policy.Current" class="tag is-success">Will be created!</span>
                  </div>
                </div>
              </nav>

              <div class="columns">
                <div v-if="policy.Current" class="column">
                  <article class="message is-primary" :class="policy.New ? '' : 'is-danger'">
                    <div class="message-header">
                      Current policy rules
                    </div>
                    <pre v-highlightjs="policy.Current"><code class="javascript"></code></pre>
                  </article>
                </div>

                <div v-if="policy.New" class="column">
                  <article class="message is-info" :class="policy.Current ? '' : 'is-success'">
                    <div class="message-header">
                    Proposed policy rules
                    </div>
                    <pre v-highlightjs="policy.New"><code class="javascript"></code></pre>
                  </article>
                </div>
              </div>
            </div>

          </article>

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
      csrf: '',
      searchString: '',
      searchType: 'changeid',
      request: null,
      bConfirm: false,
      bReject: false,
      unsealToken: '',
      progress: 0,
      required: 0
    }
  },

  mounted: function () {
  },

  computed: {
    searchURL: function () {
      var url = '/api/policy/request?type=' + this.searchType
      if (this.searchType === 'changeid') {
        url += '&id=' + this.searchString
      } else if (this.searchType === 'commit') {
        url += '&sha=' + this.searchString
      }
      return url
    },

    updateURL: function () {
      var url = '/api/policy/request/update?type=' + this.searchType
      if (this.searchType === 'changeid') {
        url += '&id=' + this.searchString
      } else if (this.searchType === 'commit') {
        url += '&sha=' + this.searchString
      }
      return url
    }
  },

  methods: {
    search: function () {
      if (this.request !== null) {
        return
      }
      this.$http.get(this.searchURL).then((response) => {
        this.csrf = response.headers['x-csrf-token']
        this.request = response.data.result
        this.progress = response.data.progress
        this.required = response.data.required
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    approve: function () {
      this.$http.post(this.updateURL, querystring.stringify({
        unseal: this.unsealToken
      }), {
        headers: {'X-CSRF-Token': this.csrf}
      })

      .then((response) => {
        this.unsealToken = ''

        // if more unseals are needed
        if (response.data.progress) {
          this.progress = response.data.progress
          this.$notify({
            title: 'Progress',
            message: response.data.progress.toString() + ' unseal tokens received so far',
            type: 'success'
          })
          if (response.data.progress === 1) {
            this.$notify({
              title: 'Timer started',
              message: 'Other operators have a 1 hour window to enter their unseal tokens',
              duration: 20000,
              type: 'warning'
            })
          }

        // if change was successfully completed
        } else {
          this.progress = this.required
          if (this.searchType === 'changeid') {
            this.request.Current = response.data.result || this.request.Current
          }
          this.$notify({
            title: 'Change success',
            message: 'Root token generated and revoked',
            type: 'success'
          })
        }
      })
      .catch((error) => {
        this.unsealToken = ''
        this.progress = 0
        this.$onError(error)
      })
    },

    reject: function () {
      this.$http.delete('/api/policy/request/' + this.searchString, {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.$notify({
          title: 'Deleted',
          message: 'Request data purged',
          type: 'warning'
        })
        this.request = null
      })
      .catch((error) => {
        this.$onError(error)
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
