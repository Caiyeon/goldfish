<template>
  <div>
    <div class="tile is-ancestor">
      <div class="tile is-parent">
        <article class="tile is-child box">

        <div class="columns">

        <!-- Left column (Form) -->
        <div class="column is-6">

          <!-- Role (if user can list them) -->
          <div v-if="availableRoles && availableRoles.length > 0" class="field">
            <label class="label">Load preset from role</label>
            <div class="control has-icons-right">
              <span class="select">
                <select v-model="selectedRole" @change="loadRoleDetails(selectedRole)">
                  <option v-for="role in availableRoles">
                    {{role}}
                  </option>
                </select>
              </span>
            </div>
          </div>

          <!-- ID -->
          <div v-if="availablePolicies.indexOf('root') > -1" class="field">
            <label class="label">ID</label>
            <div class="control">
              <input class="input is-info" type="text" placeholder="Default will be a UUID" v-model="ID">
            </div>
            <p class="help is-info">
              Root privilege
            </p>
          </div>

          <!-- Display name -->
          <div class="field">
            <label class="label">Display Name</label>
            <div class="control">
              <input class="input" type="text" placeholder="Default will be 'token'" v-model="displayName">
              <p v-if="displayName !== ''" class="help is-info">
                Display name will be 'token-{{ displayName }}'
              </p>
            </div>
          </div>

          <!-- TTL -->
          <div class="field">
            <label class="label">TTL</label>
            <div class="control">
              <input class="input" type="text"
                placeholder="e.g. '2d 12h' or '10h 30m 20s'"
                v-model="ttl"
                :class="stringToSeconds(this.ttl) < 0 ? 'is-danger' : ''">
              <p v-if="stringToSeconds(this.ttl) < 0" class="help is-danger">
                TTL cannot be negative
              </p>
              <p v-if="stringToSeconds(this.ttl) > 0" class="help is-info">
                {{ stringToSeconds(this.ttl) }} seconds
              </p>
            </div>
          </div>

          <!-- Max_TTL -->
          <div class="field">
            <label class="label">Explicit Max TTL</label>
            <div class="control">
              <input class="input" type="text"
                placeholder="e.g. '2d 12h' or '10h 30m 20s'"
                v-model="max_ttl"
                :class="stringToSeconds(this.max_ttl) < 0 ? 'is-danger' : ''">
              <p v-if="stringToSeconds(this.max_ttl) < 0" class="help is-danger">
                TTL cannot be negative
              </p>
              <p v-if="stringToSeconds(this.max_ttl) > 0" class="help is-info">
                {{ stringToSeconds(this.max_ttl) }} seconds
              </p>
            </div>
          </div>

          <!-- Renewable -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Renewable?</label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <vb-switch type="info" :checked="bRenewable" v-model="bRenewable"></vb-switch>
                </div>
              </div>
            </div>
          </div>

          <!-- Wrapping -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">Wrapped?</label>
            </div>
            <div class="field-body">
              <div class="field is-grouped">
                <div class="control">
                  <vb-switch type="info" :checked="bWrapped" v-model="bWrapped"></vb-switch>
                </div>
              </div>
              <div v-if="bWrapped" class="field">
                <input class="input" type="text"
                  placeholder="Wrap-ttl e.g. '5m'"
                  v-model="wrap_ttl"
                  :class="stringToSeconds(this.wrap_ttl) < 0 ? 'is-danger' : ''">
                <p v-if="stringToSeconds(this.wrap_ttl) < 0" class="help is-danger">
                  TTL cannot be negative
                </p>
                <p v-if="stringToSeconds(this.wrap_ttl) > 0" class="help is-info">
                  {{ stringToSeconds(this.wrap_ttl) }} seconds
                </p>
              </div>
            </div>
          </div>

          <!-- No-parent -->
          <div v-if="availablePolicies.indexOf('root') > -1" class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">
                No parent?
              </label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <vb-switch type="danger" :checked="bNoParent" v-model="bNoParent"></vb-switch>
                </div>
              </div>
            </div>
          </div>

          <!-- Period -->
          <div v-if="availablePolicies.indexOf('root') > -1" class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">
                Periodic?
              </label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <vb-switch type="danger" :checked="bPeriodic" v-model="bPeriodic"></vb-switch>
                </div>
              </div>
            </div>
          </div>
          <div v-if="availablePolicies.indexOf('root') > -1 && this.bPeriodic" class="field">
            <label class="label">Period TTL</label>
            <div class="control">
              <input class="input is-danger" type="text"
                placeholder="e.g. '2d 12h' or '10h 30m 20s'"
                v-model="period_ttl"
                :class="stringToSeconds(this.period_ttl) < 0 ? 'is-danger' : ''">
              <p v-if="stringToSeconds(this.period_ttl) < 0" class="help is-danger">
                TTL cannot be negative
              </p>
              <p v-if="stringToSeconds(this.period_ttl) > 0" class="help is-info">
                {{ stringToSeconds(this.period_ttl) }} seconds
              </p>
            </div>
          </div>

          <!-- Metadata -->
          <div class="field is-horizontal">
            <div class="field-label is-normal">
              <label class="label">
                Metadata?
              </label>
            </div>
            <div class="field-body">
              <div class="field">
                <div class="control">
                  <vb-switch type="info" :checked="bMetadata" v-model="bMetadata"></vb-switch>
                </div>
              </div>
            </div>
          </div>
          <div v-if="bMetadata" class="field">
            <p class="control">
              <textarea class="textarea"
              placeholder="Paste valid JSON here"
              v-model="metadata"></textarea>
            </p>
          </div>

          <!-- Policies -->
          <div class="field">
            <div class="control">
              <nav class="panel">

                <p class="panel-heading">Available Policies</p>
                <div class="panel-block">
                  <p class="control has-icons-left">
                    <input class="input is-small" type="text" placeholder="Search" v-model="policyFilter">
                    <span class="icon is-small is-left">
                      <i class="fa fa-search"></i>
                    </span>
                  </p>
                </div>
                <label
                  class="panel-block"
                  v-for="policy in filteredPolicies">
                  <input
                    type="checkbox"
                    :checked="selectedPolicies.indexOf(policy) > -1"
                    @click="toggle(policy)"
                    > {{ policy }} </label>
                </label>

                <p v-if="selectedRoleDetails" class="panel-heading">Role Allowed Policies</p>
                <label
                  class="panel-block"
                  v-for="policy in filteredRolePolicies">
                  <input
                    type="checkbox"
                    :checked="selectedPolicies.indexOf(policy) > -1"
                    @click="toggle(policy)"
                    > {{ policy }} </label>
                </label>

                <div class="panel-block">
                  <button class="button is-danger is-outlined is-fullwidth" @click="selectedPolicies = []">
                    Reset selected policies
                  </button>
                </div>

              </nav>
            </div>
          </div>

          <!-- Confirm button -->
          <div class="field">
            <div class="control">
              <button
              v-if="selectedPolicies.indexOf('root') > -1"
              class="button is-danger"
              @click="createToken()"
              :disabled="this.payloadJSON.metadata === 'INVALID JSON'">
                Create Root Token
              </button>

              <button
              v-else
              class="button is-primary"
              @click="createToken()"
              :disabled="selectedPolicies.length === 0 ||
              this.payloadJSON.metadata === 'INVALID JSON'">
                Create Token
              </button>

              <p v-if="selectedPolicies.length === 0" class="help is-danger">WARNING: No policies selected</p>
              <p v-if="selectedPolicies.indexOf('root') > -1" class="help is-danger">WARNING: Root policy is selected</p>
            </div>
          </div>

        <!-- ends column -->
        </div>

        <!-- Right column -->
        <div class="column is-6">

          <!-- Warnings -->
          <div v-if="availableRoles === null" class="field">
            <article class="message is-warning">
              <div class="message-body">
                <strong>Warning: Logged in user is not authorized to list roles</strong>
              </div>
            </article>
          </div>
          <div v-if="selectedRole" class="field">
            <article class="message is-warning">
              <div class="message-body">
                <strong>Warning: Some options may be overridden by role details</strong>
              </div>
            </article>
          </div>
          <div class="field"
          v-if="this.max_ttl != '' && (stringToSeconds(this.ttl) > stringToSeconds(this.max_ttl))">
            <article class="message is-warning">
              <div class="message-body">
                <strong>Warning: ttl is longer than max_ttl</strong>
              </div>
            </article>
          </div>

          <!-- Role details -->
          <div v-if="selectedRole && selectedRoleDetails" class="field">
            <label class="label">Selected role: {{selectedRole}}</label>
            <article class="message is-info">
              <div class="message-body" style="white-space: pre;">{{JSON.stringify(selectedRoleDetails, null, '\t')}}</div>
            </article>
          </div>

          <!-- Token creation response -->
          <div v-if="createdToken" class="field">
            <label class="label">Created token:</label>
            <article class="message is-success">
              <pre v-highlightjs="JSON.stringify(createdToken, null, '    ')"><code class="javascript"></code></pre>
            </article>
          </div>

          <!-- Payload preview -->
          <div class="field">
            <label class="label">Payload preview:</label>
            <article class="message is-primary">
              <pre v-highlightjs="JSON.stringify(payloadJSON, null, '    ')"><code class="javascript"></code></pre>
            </article>
          </div>

        <!-- ends column -->
        </div>

        </div>
        </article>
      </div>
    </div>
  </div>
</template>

<script>
import VbSwitch from 'vue-bulma-switch'

export default {
  components: {
    VbSwitch
  },

  data () {
    return {
      csrf: '',
      bRenewable: true,
      bNoParent: false,
      bPeriodic: false,
      bRole: false,
      bWrapped: false,
      bMetadata: false,
      ID: '',
      displayName: '',
      ttl: '',
      max_ttl: '',
      wrap_ttl: '',
      metadata: '',
      availablePolicies: ['default'],
      selectedPolicies: ['default'],
      policyFilter: '',
      num_uses: 0,
      period_ttl: '',
      createdToken: null,
      availableRoles: [],
      selectedRole: '',
      selectedRoleDetails: '',
      selectedRoleLoading: false
    }
  },

  computed: {
    // returns all policies in availablePolicies that contain the policyFilter substring
    filteredPolicies: function () {
      var filter = this.policyFilter
      return this.availablePolicies.filter(
        function (policy) {
          return policy.includes(filter)
        }
      )
    },

    filteredRolePolicies: function () {
      var filter = this.policyFilter
      if (this.selectedRoleDetails) {
        return this.selectedRoleDetails['allowed_policies'].filter(
          function (policy) {
            return policy.includes(filter)
          }
        )
      }
    },

    // returns valid JSON if metadata is set. Otherwise return null
    metadataJSON: function () {
      try {
        var json = JSON.parse(this.metadata)
        return (typeof json === 'object' && json != null) ? json : null
      } catch (e) {
        return null
      }
    },

    // constructs the JSON payload that needs to be sent to the server
    payloadJSON: function () {
      var payload = {
        'id': this.ID,
        'display_name': this.displayName,
        'ttl': this.stringToSeconds(this.ttl).toString() + 's',
        'explicit_max_ttl': this.stringToSeconds(this.max_ttl).toString() + 's',
        'renewable': !!this.bRenewable,
        'no_parent': !!this.bNoParent,
        'period': this.bPeriodic ? this.period_ttl : '',
        'no_default_policy': this.selectedPolicies.indexOf('default') === -1,
        'policies': this.selectedPolicies
      }
      if (this.bMetadata) {
        payload['meta'] = this.metadataJSON || 'INVALID JSON'
      }
      return payload
    },

    wrapParam: function () {
      if (this.bWrapped) {
        return '&wrap-ttl=' + this.stringToSeconds(this.wrap_ttl).toString() + 's'
      }
      return ''
    }
  },

  mounted: function () {
    this.$http.get('/api/users/csrf')
    .then((response) => {
      this.csrf = response.headers['x-csrf-token']
    })
    .catch((error) => {
      this.$onError(error)
    })

    // fetch available policies
    try {
      var session = JSON.parse(window.localStorage.getItem('session'))
      if (Date.now() > Date.parse(session['cookie_expiry'])) {
        throw session
      } else {
        this.availablePolicies = session.policies
      }
    } catch (e) {
      this.$notify({
        title: 'Session not found',
        message: 'Please login',
        type: 'danger'
      })
      return
    }

    // check if roles are available to logged in user
    this.$http.get('/api/users/listroles').then((response) => {
      if (response.data.result !== null) {
        this.availableRoles = response.data.result
      }
    })
    .catch((error) => {
      // if user simply doesn't have list capability on roles
      var msg = error.response.data.error || ''
      if (msg === 'User lacks capability to list roles') {
        this.availableRoles = null
      } else {
        // handle other errors the generic way
        this.$onError(error)
      }
    })

    // if root policy, fetch all available policies from server
    if (this.availablePolicies.indexOf('root') > -1) {
      this.$http.get('/api/policy').then((response) => {
        this.availablePolicies = response.data.result
        // default policy is always an option, and the first item in list
        var i = this.availablePolicies.indexOf('default')
        if (i < 0) {
          this.availablePolicies.splice(0, 0, 'default')
        } else if (i > 0) {
          var temp = this.availablePolicies[i]
          this.availablePolicies[i] = this.availablePolicies[0]
          this.availablePolicies[0] = temp
        }
      })
      .catch((error) => {
        this.$onError(error)
      })
    }
  },

  methods: {
    stringToSeconds: function (str) {
      if (str.includes('-')) {
        return -1
      }
      var totalSeconds = 0
      var days = str.match(/(\d+)\s*d/)
      var hours = str.match(/(\d+)\s*h/)
      var minutes = str.match(/(\d+)\s*m/)
      var seconds = str.match(/(\d+)$/) || str.match(/(\d+)\s*s/)
      if (days) { totalSeconds += parseInt(days[1]) * 86400 }
      if (hours) { totalSeconds += parseInt(hours[1]) * 3600 }
      if (minutes) { totalSeconds += parseInt(minutes[1]) * 60 }
      if (seconds) { totalSeconds += parseInt(seconds[1]) }
      return totalSeconds
    },

    isValidJSON: function (str) {
      try {
        JSON.parse(str)
      } catch (e) {
        console.log(false)
        return false
      }
      console.log(true)
      return true
    },

    toggle: function (policy) {
      // if already selected, unselect the policy
      if (this.selectedPolicies.indexOf(policy) > -1) {
        this.selectedPolicies.splice(this.selectedPolicies.indexOf(policy), 1)
      } else {
        this.selectedPolicies.push(policy)
      }
    },

    createToken: function () {
      // short circuit to failure if metadata is invalid
      if (this.payloadJSON.metadata === 'INVALID JSON') {
        return
      }

      this.createdToken = null
      this.$http.post('/api/users/create?type=token' + this.wrapParam, this.payloadJSON, {
        headers: {'X-CSRF-Token': this.csrf}
      })
      .then((response) => {
        this.$notify({
          title: 'Token created!',
          message: 'Details will be only shown once!',
          type: 'success'
        })
        this.createdToken = response.data.result.auth || response.data.result.wrap_info
      })
      .catch((error) => {
        this.$onError(error)
      })
    },

    loadRoleDetails: function (rolename) {
      this.selectedRoleLoading = true
      this.selectedRoleDetails = ''
      this.$http.get('/api/users/role?rolename=' + rolename)
      .then((response) => {
        this.selectedRoleDetails = response.data.result
        this.selectedRoleLoading = false
      })
      .catch((error) => {
        this.$onError(error)
        this.selectedRoleLoading = false
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

  .switch {
    top: 7px;
  }

</style>
