<template>
  <div>
    <div class="tile is-ancestor is-vertical">

      <!-- Nav bar -->
      <div class="tile is-parent">
        <article class="tile is-child box">
          <nav class="level">

            <!-- Search by name -->
            <div class="level-left">
              <div class="level-item">
                <p class="control has-icons-left">
                  <input class="input" type="text" placeholder="Filter by policy name" v-model="nameFilter">
                  <span class="icon is-small is-left">
                    <i class="fa fa-search"></i>
                  </span>
                </p>
              </div>
            </div>

            <div class="level-item">
              <p class="subtitle is-5">
                Displaying <strong> {{ filteredPolicies.length}} </strong> out of <strong>{{ policies.length }}</strong> policies
              </p>
            </div>

            <!-- Search by content -->
            <div class="level-right">
              <div class="level-item">
                <div class="field has-addons">
                  <p class="control">
                    <span class="select">
                      <select v-model="search.regex">
                      <option v-bind:value="false">Substring</option>
                      <option v-bind:value="true">Regex</option>
                      </select>
                    </span>
                  </p>
                  <p class="control">
                    <input class="input" type="text" placeholder="Filter by policy details" v-model="search.str">
                  </p>
                  <p class="control">
                    <button class="button is-info" @click="filterByDetails()" :class="loading ? 'is-loading' : ''">
                      Search
                    </button>
                  </p>
                </div>
              </div>
            </div>

          </nav>
        </article>
      </div>

      <!-- Policies table -->
      <div class="tile is-parent is-marginless is-paddingless">
        <div class="tile is-parent is-child is-vertical is-5">
          <article class="tile is-child box">
            <div class="table-responsive">
              <table class="table is-striped is-narrow">
                <thead>
                  <tr>
                    <th></th>
                    <th>Policy Name</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(entry, index) in filteredPolicies">
                    <td width="34">
                      <span class="icon">
                      <a @click="getPolicyRules(entry)">
                        <i class="fa fa-info"></i>
                      </a>
                      </span>
                    </td>
                    <td>
                      {{ entry }}
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </article>
        </div>

      <!-- Policy details -->
        <div class="tile is-parent is-vertical">
          <article class="tile is-child box">
            <h4 class="title is-4">Policy Rules</h4>

            <div class="field">
              <p class="control">
                <textarea class="textarea" placeholder="Select a policy" v-model="policyRulesModified"></textarea>
              </p>
            </div>

            <div class="field">
              <p class="control is-pulled-right">
                <a class="button is-primary is-outlined"
                  @click="addPolicyRequest()"
                  :disabled="policyRules === policyRulesModified">
                  <span>Request changes</span>
                  <span class="icon is-small">
                    <i class="fa fa-check"></i>
                  </span>
                </a>
              </p>
            </div>

          </article>
        </div>
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
      policies: [],
      policyRules: '',
      policyRulesModified: '',
      loading: false,
      nameFilter: '',
      search: {
        str: '',
        found: [],
        searched: 0,
        regex: false
      },
      selectedPolicy: ''
    }
  },

  mounted: function () {
    this.$http.get('/api/policy').then((response) => {
      this.policies = response.data.result
      this.csrf = response.headers['x-csrf-token']
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  computed: {
    filteredPolicies: function () {
      if (this.nameFilter) {
        // filter by name
        var filter = this.nameFilter
        return this.policies.filter(
          function (policy) {
            return policy.includes(filter)
          }
        )
      }
      if (this.search.str) {
        // filter by policy details
        return this.search.found
      }
      return this.policies
    }
  },

  methods: {
    getPolicyRules: function (policyName) {
      this.policyRules = ''
      this.policyRulesModified = ''
      this.selectedPolicy = policyName
      this.$http.get('/api/policy?policy=' + policyName).then((response) => {
        this.policyRules = response.data.result
        this.policyRulesModified = this.policyRules
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    filterByDetails: function () {
      if (this.search.str === '') {
        return
      }
      this.search.found = []
      this.search.searched = 0
      this.loading = this.policies.length

      // crawl through each policy
      for (var i = 0; i < this.policies.length; i++) {
        let policyName = this.policies[i]
        this.$http.get('/api/policy?policy=' + policyName).then((response) => {
          if (this.search.regex) {
            if (response.data.result.match(this.search.str) || policyName.match(this.search.str)) {
              this.search.found.push(policyName)
            }
          } else {
            if (response.data.result.includes(this.search.str) || policyName.includes(this.search.str)) {
              this.search.found.push(policyName)
            }
          }
          this.search.searched++
          this.loading = this.loading - 1 || false
        })
        .catch((error) => {
          this.$onError(error)
          this.search.searched++
          this.loading = this.loading - 1 || false
        })
      }
    },

    addPolicyRequest: function () {
      this.$http.post('/api/policy/request?policy=' + this.selectedPolicy,
        querystring.stringify({ rules: this.policyRulesModified }),
       { headers: {'X-CSRF-Token': this.csrf} })
      .then((response) => {
        this.$message({
          message: 'Your request hash is: ' + response.data.result,
          type: 'success',
          duration: 0,
          showCloseButton: true
        })
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

  .textarea {
    height: 500px;
  }
</style>
