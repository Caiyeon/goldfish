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
                  <th></th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(entry, index) in policies">
                  <td class="is-icon">
                    <a @click="getPolicyRules(index)">
                      <i class="fa fa-info"></i>
                    </a>
                  </td>
                  <td>
                    {{ entry }}
                  </td>
                  <td class="is-icon">
                    <a @click="showDeleteModal(index)">
                      <i class="fa fa-trash-o"></i>
                    </a>
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
            <textarea class="textarea" placeholder="Select a policy" disabled>{{ policyRules }}</textarea>
          </p>
        </article>
      </div>

    </div>

    <confirmModal :visible="visibleDeleteModal" :title="deleteTitle" :info="deleteInfo" @close="closeDeleteModal" @confirmed="deletePolicy(selectedIndex)"></confirmModal>

  </div>
</template>

<script>
  import Tooltip from 'vue-bulma-tooltip'
  import Vue from 'vue'
  import Notification from 'vue-bulma-notification'
  import ConfirmModal from './modals/ConfirmModal'

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

  export default {
    components: {
      Tooltip,
      ConfirmModal
    },

    data () {
      return {
        csrf: '',
        policies: [],
        policyRules: '',
        visibleDeleteModal: false,
        selectedIndex: -1
      }
    },

    computed: {
      deleteTitle: function () {
        return 'Are you sure you want to delete policy: ' + this.policies[this.selectedIndex]
      },
      deleteInfo: function () {
        return ''
      }
    },

    filters: {
    },

    mounted: function () {
      this.$http.get('/api/policies').then(function (response) {
        this.policies = response.data.result
        this.csrf = response.headers.get('x-csrf-token')
      }, function (err) {
        console.log(err.body.error)
      })
    },

    methods: {
      getPolicyRules: function (index) {
        this.policyRules = ''
        this.$http.get('/api/policies/' + this.policies[index]).then(function (response) {
          this.policyRules = response.data.result
        }, function (err) {
          console.log(err.body.error)
        })
      },

      deletePolicy: function (index) {
        this.$http.delete('/api/policies/' + this.policies[index], { headers: {'X-CSRF-Token': this.csrf} }).then(function (response) {
          this.visibleDeleteModal = false
          this.policies.splice(index, 1)
          openNotification({
            title: 'Deletion successful',
            message: '',
            type: 'success'
          })
          console.log(response.data.status)
        }, function (err) {
          console.log(err.body.error)
        })
      },

      showDeleteModal: function (index) {
        this.selectedIndex = index
        this.visibleDeleteModal = true
      },
      closeDeleteModal: function () {
        this.selectedIndex = -1
        this.visibleDeleteModal = false
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
