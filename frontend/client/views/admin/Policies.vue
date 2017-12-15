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
                  <p class="control has-icons-right">
                    <span class="select">
                      <select v-model="search.regex">
                      <option v-bind:value="false">Smart Search</option>
                      <option v-bind:value="true">Regex</option>
                      </select>
                    </span>
                  </p>
                  <p class="control">
                    <input class="input" type="text"
                    :placeholder ="search.regex ?
                      'Filter by policy details' :
                      'e.g. \'secret/foo/bar\''"
                    v-model="search.str"
                    @keyup.enter="search.regex ? filterPoliciesByRegex() : filterPoliciesByPath()">
                  </p>
                  <p class="control">
                    <button class="button is-info"
                    @click="search.regex ? filterPoliciesByRegex() : filterPoliciesByPath()"
                    :class="loading ? 'is-loading' : ''">
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
          <article class="tile is-child box is-clearfix">
            <table class="table is-fullwidth is-striped is-narrow">
              <thead>
                <tr>
                  <th>Policy Name</th>
                  <td v-if="search.searched">Capabilities</td>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in filteredPolicies">
                  <td>
                    <a class="tag is-primary"
                    :class="entry === 'root' ? 'is-danger': ''"
                    @click="getPolicyRules(entry)">
                      {{entry}}</a>
                  </td>
                  <td v-if="search.searched">
                    <div v-if="search.found[entry]" class="tags">
                      <span
                        class="tag is-rounded"
                        v-for="capability in search.found[entry]"
                        :class="colorCode(capability)"
                      >
                        {{capability}}
                      </span>
                    </div>
                  </td>
                </tr>
                <tr v-if="bNewPolicy">
                  <td>
                    <input v-if="bNewPolicy && !newPolicyName"
                      v-model.lazy="newPolicyName"
                      placeholder="Enter a policy name"></input>
                    <div v-if="bNewPolicy && newPolicyName" class="tags has-addons is-marginless is-paddingless">
                      <span class="tag is-info">
                        {{newPolicyName}}
                      </span>
                      <a class="tag is-delete" @click="bNewPolicy = false; newPolicyName = ''"></a>
                    </div>
                    <p v-if="bNewPolicy && newPolicyName && (policies.indexOf(newPolicyName) === -1)" class="help is-info">
                      You are requesting this policy to be created
                    </p>
                    <p v-if="bNewPolicy && (policies.indexOf(newPolicyName) > -1)" class="help is-danger">
                      This policy already exists! <br>Cancel this policy, click on the existing one, and then request changes
                    </p>
                  </td>
                </tr>
              </tbody>
            </table>
            <a v-if="!newPolicyName"
            class="button is-primary is-outlined is-pulled-right"
            @click="bNewPolicy = true; reset()">
              Request a new policy</a>
          </article>
        </div>

        <!-- Policy details -->
        <div class="tile is-parent is-vertical">
          <article class="tile is-child box">
            <p class="subtitle is-4">Policy Rules</p>

            <div class="field">
              <p class="control">
                <textarea class="textarea"
                placeholder="Select a policy"
                v-model="policyRulesModified"
                rows="20"></textarea>
              </p>
            </div>

            <div class="field is-grouped is-pulled-right">
              <p v-if="newPolicyName === ''" class="control">
                <a class="button is-danger is-outlined"
                  @click="addPolicyRemoveRequest()"
                  :disabled="selectedPolicy === ''">
                  <span>Request deletion</span>
                </a>
              </p>

              <p class="control">
                <a class="button is-primary is-outlined"
                  @click="addPolicyRequest()"
                  :class="newPolicyName ? 'is-info' : ''"
                  :disabled="policyRules === policyRulesModified
                  || (policies.indexOf(newPolicyName) > -1)
                  || policyRulesModified === ''">
                  <span>Request {{newPolicyName ? 'creation' : 'changes'}}</span>
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
export default {
  data () {
    return {
      bNewPolicy: false,
      policies: [],
      policyRules: '',
      policyRulesModified: '',
      loading: false,
      nameFilter: '',
      search: {
        str: '',
        found: {},
        searched: 0,
        regex: false
      },
      selectedPolicy: '',
      newPolicyName: ''
    }
  },

  mounted: function () {
    this.$http.get('/v1/policy', {
      headers: {'X-Vault-Token': this.session ? this.session.token : ''}
    })
    .then((response) => {
      this.policies = response.data.result
    })
    .catch((error) => {
      this.$onError(error)
    })
  },

  computed: {
    session: function () {
      return this.$store.getters.session
    },
    filteredPolicies: function () {
      let policies = this.policies

      // if a search has been done, only filter searched policies
      if (this.search.searched) {
        policies = Object.keys(this.search.found)
      }

      // if a plain filter is active, further filter policies
      if (this.nameFilter !== '') {
        let filter = this.nameFilter
        policies = policies.filter(
          function (policy) {
            return policy.includes(filter)
          }
        )
      }

      return policies
    }
  },

  methods: {
    reset: function () {
      this.selectedPolicy = ''
      this.policyRules = ''
      this.policyRulesModified = ''
    },

    getPolicyRules: function (policyName) {
      if (this.newPolicyName) {
        this.$notify({
          title: 'New policy in edit',
          message: 'Cancel the new policy before viewing another',
          type: 'warning'
        })
        return
      }
      this.policyRules = ''
      this.policyRulesModified = ''
      this.selectedPolicy = policyName
      this.$http.get('/v1/policy?policy=' + policyName, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      }).then((response) => {
        this.policyRules = response.data.result
        this.policyRulesModified = this.policyRules
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    filterPoliciesByRegex: function () {
      if (this.search.str === '') {
        return
      }

      // ensure the search string is valid regex
      try {
        RegExp(this.search.str)
      } catch (e) {
        this.$notify({
          title: 'Error',
          message: 'Not a valid regex string!',
          type: 'warning'
        })
        return
      }

      this.search.found = {}
      this.search.searched = 0
      this.loading = this.policies.length

      // crawl through each policy
      for (const policy of this.policies) {
        this.$http.get('/v1/policy?policy=' + encodeURIComponent(policy), {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          if (response.data.result.match(this.search.str)) {
            this.search.found[policy] = []
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

    filterPoliciesByPath: function () {
      if (this.search.str === '') {
        return
      }
      this.search.found = {}
      this.search.searched = 0
      this.loading = this.policies.length

      // for each policy, check capabilities on path (i.e. search string)
      for (const policy of this.policies) {
        this.$http.get('/v1/policy-capabilities?policy=' + encodeURIComponent(policy) +
        '&path=' + encodeURIComponent(this.search.str), {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          // add any non-'deny' policies to the found list
          if (response.data.result[0] !== 'deny') {
            this.search.found[policy] = response.data.result
          }
          this.search.searched++
          this.loading = this.loading - 1 || false
        })
        .catch((error) => {
          // notify user of any errors
          this.$onError(error)
          this.search.searched++
          this.loading = this.loading - 1 || false
        })
      }
    },

    colorCode: function (capability) {
      switch (capability) {
        case 'root':
        case 'sudo':
          return 'is-danger'

        case 'read':
        case 'list':
          return 'is-primary'

        case 'update':
        case 'create':
        case 'delete':
          return 'is-info'

        default:
          return 'is-warning'
      }
    },

    addPolicyRequest: function () {
      if (this.policyRules === this.policyRulesModified || this.policyRulesModified === '') {
        return
      }

      if (this.policies.indexOf(this.newPolicyName) > -1) {
        this.$notify({
          title: 'Policy already exists',
          message: 'Cancel this new policy, select the existing one, and edit that instead',
          type: 'warning'
        })
        return
      }

      this.$http.post('/v1/request/add', {
        type: 'policy',
        policyname: this.selectedPolicy || this.newPolicyName,
        rules: this.policyRulesModified
      }, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.$message({
          message: 'Your request ID is: ' + response.data.result,
          type: 'success',
          duration: 0,
          showCloseButton: true
        })
        if (response.data.error !== '') {
          this.$notify({
            title: 'Slack webhook',
            message: response.data.error,
            type: 'warning'
          })
        }
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    addPolicyRemoveRequest: function () {
      if (this.selectedPolicy === '' || this.newPolicyName !== '') {
        return
      }
      this.$http.post('/v1/request/add', {
        type: 'policy',
        policyname: this.selectedPolicy,
        rules: ''
      }, {
        headers: {'X-Vault-Token': this.session ? this.session.token : ''}
      })
      .then((response) => {
        this.$message({
          message: 'Your request ID is: ' + response.data.result,
          type: 'success',
          duration: 0,
          showCloseButton: true
        })
        if (response.data.error !== '') {
          this.$notify({
            title: 'Slack webhook',
            message: response.data.error,
            type: 'warning'
          })
        }
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
