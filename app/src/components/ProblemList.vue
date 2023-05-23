<template>
  <div>
      <b-table striped hover :items="message.data" :fields="fields" responsive="sm">
        <template #cell(actions)="row">
          <b-button size="sm"  @click="info(row.item.Question)" class="mr-2">
            View
          </b-button>
          <b-button size="sm" :disabled="row.item.ProblemStatus === 0" @click="showConfirmBox(row.item.ProblemID)" class="mr-2">
            Unpublish
          </b-button>
        </template>
      </b-table>

      <b-modal :id="infoModal.id" :title="infoModal.title" ok-only>
        <pre>{{ infoModal.quesiton }}</pre>
      </b-modal>
  </div>
</template>

<script>
import * as Config from '../config'
import moment from 'moment'

export default {
  name: 'ProblemList',
  data: () => ({
    message: '',
    selectedProb: '',
    infoModal: {
      id: 'info-modal',
      title: 'Problem Detail',
      quesiton: ''
    },
    fields: [
      { key: 'ProblemID', label: 'ProblemID' },
      { key: 'Ungraded', label: 'Ungraded' },
      { key: 'Correct', label: 'Correct' },
      { key: 'Incorrect', label: 'InCorrect' },
      { key: 'PublishedDate',
        label: 'Published Date',
        formatter: value => {
          return moment(value, 'YYYY-MM-DD hh:mm').format('LLL')
        }
      },
      { key: 'ProblemStatus',
        label: 'Active',
        formatter: value => {
          return value ? 'True' : 'False'
        }
      },
      { key: 'UnpublishedDate',
        label: 'UnPublished Date',
        formatter: value => {
          return moment(value, 'YYYY-MM-DD hh:mm').format('LLL')
        }
      },
      'actions'
    ]
  }),
  methods: {
    selectProblem (item) {
      console.log(item)
      this.selectedProb = item
    },
    info (item, button) {
      this.infoModal.quesiton = item
      this.$root.$emit('bv::show::modal', this.infoModal.id, button)
    },
    toast (msg) {
      this.$bvToast.toast(`${msg}`, {
        title: `Notification`,
        toaster: 'b-toaster-top-center',
        variant: 'secondary',
        solid: true
      })
    },
    showConfirmBox (id) {
      this.boxOne = ''
      this.$bvModal.msgBoxConfirm('Are you sure you want to unpublish this problem?', {
        title: 'Please Confirm',
        size: 'sm',
        buttonSize: 'sm',
        okVariant: 'danger',
        okTitle: 'YES',
        cancelTitle: 'NO',
        footerClass: 'p-2',
        hideHeaderClose: false,
        centered: true
      })
        .then(value => {
          this.boxOne = value
          console.log('Value: ', this.boxOne)
          if (this.boxOne === true) {
            this.unpublish(id)
          }
        })
        .catch(err => {
          // An error occurred
          console.log('Error: ', err)
        })
    },
    unpublish (id) {
      this.$http.delete(Config.apiUrl + '/problems/delete', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {problem_id: id}
      })
        .then(() => {
          this.toast('Problem with id ' + id + ' is unpublished.')
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    getProblemList: function () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      this.$http.get(Config.apiUrl + '/problems/status', config)
        .then((response) => {
          console.log(response)
          this.message = response.data
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    }
  },
  created: function () {
    this.getProblemList()
  }
}
</script>
<style>

</style>
