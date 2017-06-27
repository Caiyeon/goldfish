<template>
  <div>
    <div class="tile is-ancestor">
    <div class="tile is-parent">
    <article class="tile is-child box">

    <div class="columns">

      <!-- left side: form input -->
      <div class="column is-6">

        <!-- select resource type -->
        <div class="field has-addons">
          <div class="field">
            <p class="control has-icons-right">
              <span class="select">
                <select v-model="resourceType" required
                :disabled="loading || resourceType !== ''">
                  <option value="" selected hidden>
                    Select a resource type...</option>
                  <option v-for="res in supportedResourceTypes">
                    {{res}}</option>
                </select>
              </span>
            </p>
          </div>
        </div>

        <!-- resourceType: policy -->
        <div v-if="resourceType === 'Policy'">
          <label class="label">Select policy</label>
          <div class="field has-addons">
            <p v-if="policies.length === 0 && !confirmed" class="control">
              <a class="button is-primary" @click="listPolicies()" :disabled="loading">
                List policies
              </a>
            </p>

            <p v-if="policies.length === 0" class="control" :disabled="loading">
              <input class="input" type="text"
              placeholder="Or enter a policy name..."
              @keyup.enter="confirmed = true"
              :disabled="confirmed"
              v-model="selectedPolicy">
            </p>

            <p v-if="policies.length > 0"
            class="control has-icons-right"
            :disabled="loading">
              <span class="select">
                <select v-model="selectedPolicy" required
                :disabled="loading || confirmed">
                  <option value="" disabled selected hidden>
                    Select a policy...</option>
                  <option v-for="policy in policies">
                    {{policy}}
                  </option>
                </select>
              </span>
            </p>

            <p class="control">
              <a class="button is-info"
              @click="confirmed = true"
              :disabled="loading || selectedPolicy === '' || confirmed">
                Confirm
              </a>
            </p>
          </div>

          <article v-if="confirmed">
            <a class="button is-primary is-outlined"
            @click="searchDependencies('Policy', 'Tokens')"
            :class="(resultNames.includes('Tokens') && findResult('Tokens').Loading) ? 'is-loading' : ''">
              Check all current tokens
            </a>
          </article>

          <article v-if="confirmed">
            <a class="button is-primary is-outlined"
            @click="searchDependencies('Policy', 'Users')"
            :class="(resultNames.includes('Users') && findResult('Users').Loading) ? 'is-loading' : ''">
              Check all userpass users
            </a>
          </article>

          <article v-if="confirmed">
            <a class="button is-primary is-outlined"
            @click="searchDependencies('Policy', 'Roles')"
            :class="(resultNames.includes('Roles') && findResult('Roles').Loading) ? 'is-loading' : ''">
              Check all roles for allowed policies
            </a>
          </article>

          <article v-if="confirmed">
            <a class="button is-primary is-outlined"
            @click="searchDependencies('Policy', 'Approles')"
            :class="(resultNames.includes('Approles') && findResult('Approles').Loading) ? 'is-loading' : ''">
              Check all approles
            </a>
          </article>
        </div>

      <!-- end left side -->
      </div>

      <!-- right side: results -->
      <div class="column is-6">
        <div v-for="result in results" class="field">
          <article v-if="!result.Loading && !result.Dependents.length" class="message is-success">
            <div class="message-body">
              <strong>No dependent {{result.Type | lowercase}} found</strong>
            </div>
          </article>
          <article v-if="!result.Loading && result.Dependents.length" class="message is-warning">
            <div class="message-body">
              <strong>{{result.Subtype}} of dependent {{result.Type | lowercase}}: </strong>
              <ul>
                <li v-for="dep in result.Dependents">{{dep}}</li>
              </ul>
            </div>
          </article>
        </div>
      </div>

    </div>
    </article>
    </div>
    </div>
  </div>
</template>

<script>
export default {
  data () {
    return {
      csrf: '',
      supportedResourceTypes: ['Policy'],
      resourceType: '',
      loading: false,
      confirmed: false,
      results: [],

      policies: [],
      selectedPolicy: ''
    }
  },

  mounted: function () {
  },

  filters: {
    lowercase: function (value) {
      return value ? value.toString().toLowerCase() : ''
    }
  },

  computed: {
    resultNames: function () {
      var names = []
      for (var i = 0; i < this.results.length; i++) {
        names.push(this.results[i].Type)
      }
      return names
    }
  },

  methods: {
    // results is an array rather than map because arrays are reactive
    // but purging a particular item from an array requires some loops
    purgeResult: function (key) {
      function matchesKey (result) {
        return (result && result.Type === key)
      }
      let index = this.results.findIndex(matchesKey)
      if (index !== -1) {
        this.results.splice(index, 1)
      }
    },

    findResult: function (key) {
      function matchesKey (result) {
        return (result && result.Type === key)
      }
      let index = this.results.findIndex(matchesKey)
      if (index !== -1) {
        return this.results[index]
      }
    },

    searchDependencies: function (resourceType, searchType) {
      // route to appropriate parsing function
      if (resourceType === 'Policy') {
        switch (searchType) {
          case 'Tokens':
            return this.searchPolicyTokens()
          case 'Users':
            return this.searchPolicyUsers()
          case 'Roles':
            return this.searchPolicyRoles()
          case 'Approles':
            return this.searchPolicyApproles()
          default:
            this.$notify({
              title: 'Not supported',
              message: 'Search type ' + this.searchType + ' is not supported',
              type: 'warning'
            })
        }
      } else {
        this.$notify({
          title: 'Not supported',
          message: 'Resource type ' + this.resourceType + ' is not supported',
          type: 'warning'
        })
      }
    },

    listPolicies: function () {
      // if page is loading, this is disabled
      if (this.loading === true) {
        return
      }
      this.loading = true
      this.selectedPolicy = ''

      // fetch list of policies and set csrf
      this.$http.get('/api/policy').then((response) => {
        this.policies = response.data.result
        this.csrf = response.headers['x-csrf-token']
        this.loading = false
      })
      .catch((error) => {
        this.$onError(error)
        this.loading = false
      })
    },

    // search functions for resource 'Policy'
    searchPolicyTokens: function () {
      // remove previously fetched result if it exists
      this.purgeResult('Tokens')
      var result = {
        Type: 'Tokens',
        Loading: 1,
        Subtype: 'Accessors',
        Dependents: []
      }
      this.results.push(result)
      let policy = this.selectedPolicy

      // for each 300 tokens, fetch, filter, and push to result
      this.$http.get('/api/tokencount').then((response) => {
        result.Loading = Math.ceil(response.data.result / 300)

        for (var i = 0; i < Math.ceil(response.data.result / 300); i++) {
          this.$http.get('/api/users?type=token&offset=' + (i * 300).toString())
          .then((response) => {
            for (var j = 0; j < response.data.result.length; j++) {
              if (response.data.result[j].policies.findIndex(function (p) { return (p === policy) }) > -1) {
                result.Dependents.push(response.data.result[j].accessor)
              }
            }
            result.Loading = result.Loading - 1 || false
          })
          .catch((error) => {
            this.$onError(error)
            result.Loading = result.Loading - 1 || false
          })
        }
      })
      .catch((error) => {
        this.$onError(error)
      })
    },
    searchPolicyUsers: function () {
      // remove previously fetched result if it exists
      this.purgeResult('Users')
      var result = {
        Type: 'Users',
        Loading: 1,
        Subtype: 'Usernames',
        Dependents: []
      }
      this.results.push(result)

      // fetch all users and filter by policy
      this.$http.get('/api/users?type=userpass').then((response) => {
        let users = response.data.result
        for (var i = 0; i < users.length; i++) {
          if (users[i].Policies.includes(this.selectedPolicy)) {
            result.Dependents.push(users[i].Name)
          }
        }
        result.Loading = false
      })
      .catch((error) => {
        this.$onError(error)
        this.purgeResult('Users')
      })
    },
    searchPolicyRoles: function () {
      // remove previously fetched result if it exists
      this.purgeResult('Roles')
      var result = {
        Type: 'Roles',
        Loading: 1,
        Subtype: 'Rolenames',
        Dependents: []
      }
      this.results.push(result)
      let policy = this.selectedPolicy

      // fetch all users and filter by policy
      this.$http.get('/api/users/listroles').then((response) => {
        // if there are no roles found
        if (response.data.result === null) {
          result.Loading = false
        } else {
          // otherwise, for each role, fetch allowed policies and filter
          result.Loading = response.data.result.length
          for (var i = 0; i < response.data.result.length; i++) {
            let rolename = response.data.result[i]
            this.$http.get('/api/users/role?rolename=' + rolename)
            .then((response) => {
              if (response.data.result && response.data.result['allowed_policies'].includes(policy)) {
                result.Dependents.push(rolename)
              }
              result.Loading = result.Loading - 1 || false
            })
            .catch((error) => {
              this.$onError(error)
              result.Loading = result.Loading - 1 || false
            })
          }
        }
      })
      .catch((error) => {
        this.$onError(error)
        this.purgeResult('Roles')
      })
    },
    searchPolicyApproles: function () {
      // remove previously fetched result if it exists
      this.purgeResult('Approles')
      var result = {
        Type: 'Approles',
        Loading: 1,
        Subtype: 'RoleIDs',
        Dependents: []
      }
      this.results.push(result)

      // fetch all users and filter by policy
      this.$http.get('/api/users?type=approle').then((response) => {
        let users = response.data.result
        for (var i = 0; i < users.length; i++) {
          if (users[i].Policies.includes(this.selectedPolicy)) {
            result.Dependents.push(users[i].Roleid)
          }
        }
        result.Loading = false
      })
      .catch((error) => {
        this.$onError(error)
        this.purgeResult('Approles')
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

  select:invalid {
    color: gray;
  }
</style>
