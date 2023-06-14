<template>
  <div>
      <b-table striped hover :items="message.data" :fields="fields" responsive="sm">
        <template #cell(actions)="row">
          <div class="sub-action">
          <b-button size="sm"  @click="info('Problem Description', row.item.Question)" class="mr-2">
            View Problem
          </b-button>
          <b-button size="sm" :disabled="row.item.ProblemStatus === 0" @click="showConfirmBox('unpublish', row.item)" class="mr-2">
            Unpublish Problem
          </b-button>
          </div>
          <div class="sub-action">
          <b-button size="sm" :disabled="!row.item.Solution" @click="info('Solution Code', row.item.Solution)" class="mr-2">
            View Solution
          </b-button>
          <b-button size="sm" :disabled="!row.item.Solution" @click="showConfirmBox('broadcast', row.item)" class="mr-2">
            Broadcast Solution
          </b-button>
          </div>
        </template>
      </b-table>

      <b-modal size='xl' :id="infoModal.id" :title="infoModal.title" ok-only>
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
      title: '',
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
      { key: 'LifeTime',
        label: 'Deadline',
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
    info (title, block, button) {
      this.infoModal.title = title
      this.infoModal.quesiton = block
      this.$root.$emit('bv::show::modal', this.infoModal.id, button)
    },
    toast (msg) {
      this.$bvToast.toast(`${msg}`, {
        title: `Notification`,
        toaster: 'b-toaster-top-center',
        variant: 'secondary',
        autoHideDelay: 2000,
        solid: true
      })
    },
    showConfirmBox (msg, item) {
      this.boxOne = ''
      this.$bvModal.msgBoxConfirm('Are you sure you want to ' + msg + ' this ?', {
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
            msg === 'unpublish' ? this.unpublish(item) : this.broadcast(item)
          }
        })
        .catch(err => {
          // An error occurred
          console.log('Error: ', err)
        })
    },
    unpublish (item) {
      this.$http.delete(Config.apiUrl + '/problems/delete', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {problem_id: item.ProblemID}
      })
        .then(() => {
          this.toast('Problem with id ' + item.ProblemID + ' is unpublished.')
        })
        .catch(function (error) {
          console.log(error)
        })
    },
    broadcast (item) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'solution_id': item.SolutionID
      }
      this.$http.post(Config.apiUrl + '/solution/broadcast', postBody, config)
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is on watch list.')
          this.toast('Solution for problem with id ' + item.ProblemID + ' is broadcasted.')
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
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

.sub-action {
  margin: 5px;
}

</style>
