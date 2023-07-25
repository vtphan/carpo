<template>
  <div>
      <b-table striped hover :items="message.data" :fields="fields" responsive="sm">
        <template #cell(OnWatch)="data" >
          <a href="javascript:;" @click="fetchWatch(data.item.ProblemID)">{{ data.item.OnWatch }}</a>
        </template>
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

      <b-modal size='xl' :id="watchlist.id" :title="watchlist.title" ok-only>
        <b-table striped hover :items="watchlist.watch.data" responsive="sm" :fields="['SubmissionID', 'Reason', 'Action']">
        <template #cell(Action)="row">
          <div class="sub-action">
            <b-button size="sm" v-b-modal.modal-multi-3 @click="getWatchedSub(row.item.SubmissionID)" class="mr-2">
              See Code
            </b-button>
          </div>
        </template>
        </b-table>
      </b-modal>

      <b-modal id="modal-multi-3" size="xl" :hide-footer="true">
        <template #modal-title>
            On Watch Snapshot {{ timeDiff(selectedSub.created_at) }} ago
            <b-badge v-if="selectedSub.reason" variant="info">Tag: {{ selectedSub.reason }}</b-badge>
          </template>
          <codemirror v-model="selectedSub.code" :options="cmOptions" />
          <b-row>
            <b-col cols="6" >
            </b-col>
            <b-col cols="6" >
              <div style="text-align: right">
                <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.submission_id)">Send Feedback</b-button>
              </div>
            </b-col>
          </b-row>

      </b-modal>
  </div>
</template>

<script>
import * as Config from '../config'
import moment from 'moment'
import { codemirror } from 'vue-codemirror'
import 'codemirror/lib/codemirror.css'
// language
import 'codemirror/mode/python/python.js'

// theme css
import 'codemirror/theme/duotone-light.css'

export default {
  name: 'ProblemList',
  components: {
    codemirror
  },
  data: () => ({
    message: '',
    selectedProb: '',
    selectedSub: '',
    cmOptions: {
      autoRefresh: true,
      tabSize: 4,
      styleActiveLine: true,
      lineNumbers: true,
      line: true,
      mode: 'application/x-httpd-python',
      lineWrapping: true,
      theme: 'duotone-light'
    },
    infoModal: {
      id: 'info-modal',
      title: '',
      quesiton: ''
    },
    watchlist: {
      id: 'on-watch-modal',
      title: '',
      watch: ''
    },
    fields: [
      { key: 'ProblemID', label: 'ProblemID' },
      { key: 'Ungraded', label: 'Ungraded' },
      { key: 'Correct', label: 'Correct' },
      { key: 'Incorrect', label: 'InCorrect' },
      { key: 'OnWatch', label: 'OnWatch' },
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
    timeDiff (dbTimestamp) {
      return moment.duration(moment().diff(moment(dbTimestamp))).humanize()
      // https://stackoverflow.com/questions/18623783/get-the-time-difference-between-two-datetimes
    },
    fetchWatch (problemId, button) {
      console.log('Fetch watch list for this problem_id: ', problemId)
      this.watchlist.title = 'Watch messages for ProblemID ' + problemId
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'problem_id': problemId}
      }
      this.$http.get(Config.apiUrl + '/problems/on_watch', config)
        .then((response) => {
          // console.log(response)
          this.watchlist.watch = response.data
          this.$root.$emit('bv::show::modal', this.watchlist.id, button)
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
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
    sendFeedback (submission, id) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': submission.student_id,
        'submission_id': id,
        'problem_id': submission.problem_id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody, config)
        .then(data => {
          // alert('Feedback sent to student.')
          this.toast('Feedback is sent to student.')
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
    },
    getWatchedSub (subId) {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {
          'sort_by': this.sorting,
          'submission_id': subId
        }}
      this.$http.get(Config.apiUrl + '/snapshots/watch', config)
        .then((response) => {
          this.selectedSub = response.data
          // console.log('watched' + JSON.stringify(this.selectedSub))
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
