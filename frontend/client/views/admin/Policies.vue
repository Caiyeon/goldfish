<template>
  <div>
    <div class="tile is-ancestor">

      <div class="tile is-parent is-vertical is-6">
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
                <tr v-for="(entry, index) in policies">
                  <td width="34">
                    <span class="icon">
                    <a @click="getPolicyRules(index)">
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

      <div class="tile is-parent is-vertical is-6">
        <article class="tile is-child box">
          <h4 class="title is-4">Policy Rules</h4>
          <p class="control">
            <textarea class="textarea" placeholder="Select a policy" v-model="policyRules"></textarea>
          </p>
        </article>
      </div>

    </div>

  </div>
</template>

<script>
  import Vue from 'vue'
  import Notification from 'vue-bulma-notification'

  const NotificationComponent = Vue.extend(Notification)

  const openNotification = (propsData = {
    title: '',
    message: '',
    type: '',
    direction: '',
    duration: 4500,
    container: '.notifications'
  }) => {
    return new NotificationComponent({
      el: document.createElement('div'),
      propsData
    })
  }

  function handleError (error) {
    if (error.response.data.error) {
      openNotification({
        title: 'Error: ' + error.response.status,
        message: error.response.data.error,
        type: 'danger'
      })
      console.log(error.response.data.error)
    } else {
      openNotification({
        title: 'Error',
        message: 'Please login first',
        type: 'danger'
      })
      console.log(error.message)
    }
  }

  export default {
    components: {
    },

    data () {
      return {
        policies: [],
        policyRules: '',
        selectedIndex: -1
      }
    },

    computed: {
    },

    filters: {
    },

    mounted: function () {
      this.$http.get('/api/policies')
        .then((response) => {
          this.policies = response.data.result
        })
        .catch((error) => {
          handleError(error)
        })
    },

    methods: {
      getPolicyRules: function (index) {
        this.policyRules = ''
        this.$http.get('/api/policies/' + this.policies[index])
          .then((response) => {
            this.policyRules = response.data.result
          })
          .catch((error) => {
            handleError(error)
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
