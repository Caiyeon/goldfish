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
                      'foo/bar matches foo/*'"
                    v-model="search.str"
                    @keyup.enter="filterByDetails()">
                  </p>
                  <p class="control">
                    <button class="button is-info"
                    @click="filterByDetails()"
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
        found: [],
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
      if (this.nameFilter) {
        // filter by name
        var filter = this.nameFilter
        return this.policies.filter(
          function (policy) {
            return policy.includes(filter)
          }
        )
      }
      if (this.search.searched) {
        // filter by policy details
        return this.search.found
      }
      return this.policies
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
        this.$http.get('/v1/policy?policy=' + policyName, {
          headers: {'X-Vault-Token': this.session ? this.session.token : ''}
        }).then((response) => {
          var searchString = this.search.regex ? this.search.str : this.makeRegex(this.search.str)
          if (response.data.result.match(searchString)) {
            this.search.found.push(policyName)
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
    },

    makeRegex: function (str) {
      var lastSlash = str.lastIndexOf('/')
      if (lastSlash === -1) {
        // if slash doesn't exist, match 'foo', 'foo*', 'fo*', 'f*', '*'
        var lastWord = str
        var returnString = '"(' + str
        for (var i = str.length + 1; i > 0; i--) {
          returnString += '|' + lastWord + '\\*'
          lastWord = lastWord.substring(0, i - 2)
        }
      } else {
      // if slash does exist, match 'foo/bar', 'foo/bar*', 'foo/ba*', 'foo/b*', 'foo/*'
        lastWord = str.substring(lastSlash + 1, str.length)
        if (lastWord === '') {
          var replaced = str.replace('/', '\\/')
          return '"(' + replaced + '|' + replaced + '\\*)"'
        }
        returnString = '"' + str.substring(0, lastSlash)
        returnString = returnString.replace('/', '\\/') + '\\/(' + lastWord
        for (i = lastWord.length + 1; i > 0; i--) {
          returnString += '|' + lastWord + '\\*'
          lastWord = lastWord.substring(0, i - 2)
        }
      }
      // prefix and suffix return with double quotes to ensure it matches the full path only
      return returnString + ')"'
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
