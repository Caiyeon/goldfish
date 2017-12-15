pre class="is-paddingless" v-highlightjs<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <!-- Search bar -->
          <div class="field">
            <label class="label">Enter a request ID or github commit hash:</label>
            <div class="field has-addons">
              <p v-if="request !== null" class="control">
                <a class="button is-danger" @click="reset()">
                  Reset
                </a>
              </p>
              <p class="control">
                <input v-focus class="input" type="text"
                placeholder="Enter a request ID"
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
            <div class="field">
              <div class="control">
                <label class="radio">
                  <input type="radio" v-model="show" value="syntax">
                  Highlight syntax
                </label>
                <label class="radio">
                  <input type="radio" v-model="show" value="diffSingle">
                  Line-by-line diff
                </label>
                <label class="radio">
                  <input type="radio" v-model="show" value="diffDouble">
                  Side-by-side diff
                </label>
              </div>
            </div>
          </div>

          <!-- Request type: policy -->
          <article v-if="request !== null && request['Type'] === 'policy'">
            <br>
            <!-- Request general info -->
            <article class="message is-primary">
              <div class="message-body">
                <strong>Request type: </strong>{{request.Type}}<br>
                <strong>Policy name: </strong>{{request.PolicyName}}<br>
                <strong>Requester display name: </strong>{{request.Requester}}<br>
                <strong>Requester accessor hash: </strong>{{request.RequesterHash}}<br>
                <strong>Unseal progress: </strong>{{request.Progress}} out of {{request.Required}}
                  <strong>{{request.Progress === request.Required ? ' Done!' : ''}}</strong>
              </div>
            </article>

            <!-- Approve/Reject -->
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
                  <input class="input" type="password"
                  placeholder="Enter an unseal key"
                  v-model="unsealKey"
                  @keyup.enter="approve()">
                </p>
                <p class="control">
                  <a class="button is-info" @click="approve()">
                    Confirm
                  </a>
                </p>
              </div>
            </div>

            <!-- syntax-highlighted diff -->
            <div v-if="show === 'syntax'" class="columns">
              <div v-if="request.Previous" class="column">
                <article class="message is-primary" :class="request.Proposed ? '' : 'is-danger'">
                  <div class="message-header">
                    {{request.Proposed ? 'Current policy rules' : 'Will be deleted!'}}
                  </div>
                  <pre class="is-paddingless" v-highlightjs="request.Previous"><code class="ruby"></code></pre>
                </article>
              </div>

              <div v-if="request.Proposed" class="column">
                <article class="message is-info" :class="request.Previous ? '' : 'is-success'">
                  <div class="message-header">
                  {{request.Previous ? 'Proposed policy rules' : 'Will be created!'}}
                  </div>
                  <pre  class="is-paddingless" v-highlightjs="request.Proposed"><code class="ruby"></code></pre>
                </article>
              </div>
            </div>

            <!-- diff via jsdiff and diff2html -->
            <div v-if="show === 'diffSingle' || show === 'diffDouble'" v-html="diff"></div>
          </article>

          <!-- Request type: github -->
          <article v-if="request !== null && request['Type'] === 'github'">
            <br>
            <!-- Request general info -->
            <article class="message is-primary">
              <div class="message-body">
                <strong>Request type: </strong>{{request.Type}}<br>
                <strong>Github commit hash: </strong>{{request.CommitHash}}<br>
                <strong>Number of policies affected: </strong>{{Object.keys(request.Changes).length}}<br>
                <strong>Requester display name: </strong>{{request.Requester}}<br>
                <strong>Requester accessor hash: </strong>{{request.RequesterHash}}<br>
                <strong>Unseal progress: </strong>{{request.Progress}} out of {{request.Required}}
                  <strong>{{request.Progress === request.Required ? ' Done!' : ''}}</strong>
              </div>
            </article>

            <!-- Approve/Reject -->
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
                  <input class="input" type="password"
                  placeholder="Enter an unseal key"
                  v-model="unsealKey"
                  @keyup.enter="approve()">
                </p>
                <p class="control">
                  <a class="button is-info" @click="approve()">
                    Confirm
                  </a>
                </p>
              </div>
            </div>

            <!-- syntax-highlighted diff -->
            <div class="box"
            v-if="show === 'syntax' && request.Progress !== request.Required"
            v-for="(details, policy) in request.Changes">
              <!-- policy name title and status tag -->
              <nav class="level">
                <div class="level-left">
                  <div class="level-item">
                    <div class="content is-marginless is-paddingless">
                      <h3 class="is-marginless is-paddingless">{{policy}}</h3>
                    </div>
                  </div>
                  <div class="level-item">
                    <span v-if="details.Previous && details.Proposed" class="tag is-info">Will be changed!</span>
                    <span v-if="details.Previous && !details.Proposed" class="tag is-danger">Will be deleted!</span>
                    <span v-if="!details.Previous && details.Proposed" class="tag is-success">Will be created!</span>
                  </div>
                </div>
              </nav>

              <div class="columns">
                <div v-if="details.Previous" class="column">
                  <article class="message is-primary" :class="details.Proposed ? '' : 'is-danger'">
                    <div class="message-header">
                      Current policy rules
                    </div>
                    <pre class="is-paddingless" v-highlightjs="details.Previous"><code class="ruby"></code></pre>
                  </article>
                </div>

                <div v-if="details.Proposed" class="column">
                  <article class="message is-info" :class="details.Previous ? '' : 'is-success'">
                    <div class="message-header">
                    Proposed policy rules
                    </div>
                    <pre class="is-paddingless" v-highlightjs="details.Proposed"><code class="ruby"></code></pre>
                  </article>
                </div>
              </div>
            </div>

            <!-- diff via jsdiff and diff2html -->
            <div v-if="(show === 'diffSingle' || show === 'diffDouble') && request.Progress !== request.Required" v-html="diff"></div>
          </article>

          <!-- Request type: token -->
          <article v-if="request !== null && request['Type'] === 'token'">
            <br>
            <!-- Request general info -->
            <article class="message is-primary">
              <div class="message-body">
                <strong>Request type: </strong>{{request.Type}}<br>
                <strong>Requester display name: </strong>{{request.Requester}}<br>
                <strong>Requester accessor hash: </strong>{{request.RequesterHash}}<br>
                <strong>Unseal progress: </strong>{{request.Progress}} out of {{request.Required}}
                  <strong>{{request.Progress === request.Required ? ' Done!' : ''}}</strong>
              </div>
            </article>

            <!-- Approve/Reject -->
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
                  <input class="input" type="password"
                  placeholder="Enter an unseal key"
                  v-model="unsealKey"
                  @keyup.enter="approve()">
                </p>
                <p class="control">
                  <a class="button is-info" @click="approve()">
                    Confirm
                  </a>
                </p>
              </div>
            </div>

            <!-- Request details -->
            <div class="field">
              <label class="label">Request payload:</label>
              <div class="field is-grouped is-grouped-multiline">
                <div class="control">
                  <div class="tags has-addons">
                    <span class="tag">Orphan?</span>
                    <span class="tag is-primary"
                    :class="(request.Orphan || tokenRequestPreview.no_parent) ? 'is-warning': ''">
                      {{(request.Orphan || tokenRequestPreview.no_parent) ? 'Yes' : 'No'}}
                    </span>
                  </div>
                </div>
                <div v-if="request.Role" class="control">
                  <div class="tags has-addons">
                    <span class="tag">Role</span>
                    <span class="tag is-warning">{{request.Role}}</span>
                  </div>
                </div>
                <div v-if="request.Wrap_ttl" class="control">
                  <div class="tags has-addons">
                    <span class="tag">Wrap_ttl</span>
                    <span class="tag is-primary">{{request.Wrap_ttl}}</span>
                  </div>
                </div>
              </div>
              <div class="columns">
                <div class="column">
                  <article class="message is-primary">
                    <pre class="is-paddingless" v-highlightjs="JSON.stringify(tokenRequestPreview, null, '    ')"><code class="javascript"></code></pre>
                  </article>
                </div>
                <div class="column">
                  <article v-if="request.CreateResponse" class="message is-primary">
                    <pre class="is-paddingless" v-highlightjs="JSON.stringify(request.CreateResponse.wrap_info, null, '    ')"><code class="javascript"></code></pre>
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
const jsdiff = require('diff')
const diff2html = require('diff2html').Diff2Html

export default {
  data () {
    return {
      searchString: '',
      request: null,
      bConfirm: false,
      bReject: false,
      unsealKey: '',
      show: 'syntax'
    }
  },

  mounted: function () {
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },

    // returns a constructed JSON object with the token creation request
    tokenRequestPreview: function () {
      if (!this.request || this.request['Type'] !== 'token' || !this.request['CreateRequest']) {
        return {}
      }
      return this.request.CreateRequest
    },

    diff: function () {
      if (!this.request || this.show === 'syntax') {
        return ''
      }

      let format = 'line-by-line'
      if (this.show === 'diffDouble') {
        format = 'side-by-side'
      }

      if (this.request['Previous'] && this.request['Proposed']) {
        let diff = jsdiff.createPatch(
          this.request.PolicyName,
          this.request.Previous || '',
          this.request.Proposed || '',
          '', '', {context: 10000}
        )
        return diff2html.getPrettyHtml(diff, {inputFormat: 'diff', outputFormat: format, matching: 'lines'})
      }
      if (this.request['Changes']) {
        let diff = ''
        for (const policy in this.request.Changes) {
          if (this.request.Changes.hasOwnProperty(policy)) {
            diff = diff + jsdiff.createPatch(
              policy,
              this.request.Changes[policy].Previous || '',
              this.request.Changes[policy].Proposed || '',
              '', '', {context: 10000}
            )
          }
        }
        return diff2html.getPrettyHtml(diff, {inputFormat: 'diff', outputFormat: format, matching: 'lines'})
      }
      return ''
    }
  },

  methods: {
    reset: function () {
      this.request = null
      this.bConfirm = false
      this.bReject = false
      this.unsealKey = ''
    },

    // requests for the request object from backend
    search: function () {
      // force user to use reset button if they want to find another request
      if (this.request !== null) {
        return
      }
      this.$http.get('/v1/request?hash=' + this.searchString, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      }).then((response) => {
        this.request = response.data.result
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    approve: function () {
      this.$http.post('/v1/request/approve?hash=' + this.searchString, {
        unseal: this.unsealKey
      }, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      }).then((response) => {
        this.unsealKey = ''
        this.request = response.data.result
        // notify user of progress
        if (this.request.Progress === this.request.Required) {
          this.$notify({
            title: 'Change success',
            message: 'Root token generated and revoked',
            type: 'success'
          })
          this.bConfirm = false
          this.bReject = false

          // if there is wrapped info, notify user
          this.$message({
            message: 'Created resource is in wrapping token: ' + this.request.CreateResponse.wrap_info.token,
            type: 'success',
            duration: 0,
            showCloseButton: true
          })
        } else {
          if (this.request.Progress === 1) {
            this.$notify({
              title: 'Timer started',
              message: 'Other operators have a 1 hour window to enter their unseal keys',
              duration: 20000,
              type: 'warning'
            })
          } else {
            this.$notify({
              title: 'Progress',
              message: this.request.Progress.toString() + ' unseal keys received so far',
              type: 'success'
            })
          }
        }
      })
      .catch((error) => {
        this.unsealKey = ''
        try {
          if (error.response.data.error.includes('Progress has been reset')) {
            this.request.Progress = 0
          } else if (error.response.data.error.includes('Request has been deleted')) {
            this.$notify({
              title: 'Error',
              message: 'This request contains invalid data, and has been deleted as a result',
              type: 'error'
            })
          }
          this.$onError(error)
        } catch (e) {
          this.$onError(error)
        }
      })
    },

    reject: function () {
      this.$http.delete('/v1/request/reject?hash=' + this.searchString, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.$notify({
          title: 'Deleted',
          message: 'Request data purged',
          type: 'warning'
        })
        this.reset()
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
