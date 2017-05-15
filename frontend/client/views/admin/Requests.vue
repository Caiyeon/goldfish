<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent is-vertical">
        <article class="tile is-child box">

          <div class="field">
            <!-- <label class="label">Search by change ID:</label> -->
            <div class="field has-addons">
              <p class="control">
                <input class="input" type="text"
                placeholder="Enter a change ID"
                v-model="searchString"
                @keyup.enter="search()">
              </p>
              <p class="control">
                <a class="button is-info" @click="search()">
                  Search
                </a>
              </p>
            </div>
          </div>

          <article v-if="request">
            <br>

            <article class="message is-primary">
              <div class="message-body">
                <strong>Requester: </strong>{{request.Requester}}<br>
                <strong>Policy: </strong>{{request.Policy}}<br>
                <strong>Unseals required: </strong>{{request.Required}}
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
                  <div class="message-body" style="white-space: pre;">{{request.Current}}</div>
                </article>
              </div>

              <div class="column">
                <article class="message is-info">
                  <div class="message-header">
                  Proposed policy rules
                  </div>
                  <div class="message-body" style="white-space: pre;">{{request.New}}</div>
                </article>
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
      request: null,
      changeID: '',
      bConfirm: false,
      bReject: false,
      unsealToken: ''
    }
  },

  mounted: function () {
  },

  computed: {
  },

  methods: {
    search: function () {
      this.$http.get('/api/policy/request' + '?id=' + this.searchString).then((response) => {
        this.csrf = response.headers['x-csrf-token']
        this.request = response.data.result
        this.changeID = this.searchString
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    approve: function () {
      this.$http.post('/api/policy/request/' + this.searchString, querystring.stringify({
        unseal: this.unsealToken
      }), {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.unsealToken = ''
        if (response.data.progress) {
          this.$notify({
            title: 'Progress',
            message: response.data.progress.toString() + ' unseal tokens received so far',
            type: 'success'
          })
        } else {
          this.request.Current = response.data.result || this.request.Current
          this.$notify({
            title: 'Change success',
            message: 'Root token generated and revoked',
            type: 'success'
          })
        }
      })
      .catch((error) => {
        this.unsealToken = ''
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
