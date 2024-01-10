<template>
  <div>
      <b-table striped hover :items="message.data" :fields="fields" responsive="sm">
        <template #cell(OnWatch)="data" >
          <a href="javascript:;" @click="fetchWatch(data.item.problem_id)">{{ data.item.on_watch }}</a>
        </template>
        <template #cell(actions)="row">
          <div class="sub-action">
            <b-button size="sm"  @click="info('Problem Description', row.item.problem_id, row.item.question, row.item)" class="mr-2">
              View Problem
            </b-button>
            <b-button size="sm" :disabled="row.item.status === 0" @click="showConfirmBox('unpublish', row.item)" class="mr-2">
              Unpublish Problem
            </b-button>
            </div>
            <div class="sub-action">
            <b-button size="sm" :disabled="!row.item.solution_code" @click="info('Solution Code', row.item.problem_id, row.item.solution_code, row.item)" class="mr-2">
              View Solution
            </b-button>
            <b-button size="sm" :disabled="!row.item.solution_id" @click="showConfirmBox('broadcast', row.item)" class="mr-2">
              Broadcast Solution
            </b-button>
          </div>
        </template>
      </b-table>

      <b-modal size='xl' :id="infoModal.id" :title="infoModal.title" ok-only style="background-color: #eeeee4;!important">
        <pre>{{ infoModal.quesiton }}</pre>
        <!-- Save Problem Tag [Footer] -->
        <hr style="width:100%;text-align:left;margin-left:0">
        <div v-if="infoModal.title=='Problem Description'">
          <div class="row" style="margin: 5px;">
            <h4 style="margin: 5px;">Tag: </h4>
            <multiselect style="width: 50%;" v-model="select_tag" track-by="id" label="name" placeholder="Select one" :options="available_tags.data" @select="saveProblemTag" @remove="remove_tag" :multiple="true" :close-on-select="false" :clear-on-select="false" :searchable="false">
              <template slot="singleLabel" slot-scope="{ option }"><strong>{{ option.name }}</strong> </template>
            </multiselect>
            <h5 class="new-tag-link" v-on:click="newTag()" > Create New Tag </h5>
            <!-- <b-button type="submit" variant="primary" @click="saveProblemTag(infoModal.pID)">Save Tag</b-button> -->
          </div>
          <!-- <pre class="language-json"><code>{{ select_tag  }}</code></pre> -->
        </div>
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

import Multiselect from 'vue-multiselect'

export default {
  name: 'ProblemList',
  components: {
    codemirror,
    Multiselect
  },
  data: () => ({
    message: '',
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
      quesiton: '',
      pID: ''
    },
    watchlist: {
      id: 'on-watch-modal',
      title: '',
      watch: ''
    },
    fields: [
      { key: 'problem_id', label: 'ProblemID' },
      { key: 'ungraded', label: 'Ungraded' },
      { key: 'correct', label: 'Correct' },
      { key: 'incorrect', label: 'InCorrect' },
      { key: 'on_watch', label: 'OnWatch' },
      { key: 'published_at',
        label: 'Published Date',
        formatter: value => {
          return moment(value, 'YYYY-MM-DD hh:mm').format('LLL')
        }
      },
      { key: 'status',
        label: 'Active',
        formatter: value => {
          return value ? 'True' : 'False'
        }
      },
      { key: 'lifetime',
        label: 'Deadline',
        formatter: value => {
          return moment(value, 'YYYY-MM-DD hh:mm').format('LLL')
        }
      },
      'actions'
    ],
    available_tags: '',
    select_tag: []
  }),
  methods: {
    info (title, pID, block, item, button) {
      this.infoModal.title = title
      this.infoModal.quesiton = block
      this.infoModal.pID = pID
      this.$root.$emit('bv::show::modal', this.infoModal.id, button)
      // this.getAvailableTags()
      // this.select_tag = []
      this.select_tag = item.tag
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
    newTag () {
      var link = document.createElement('a')
      link.href = window.location.origin + '/#/tags?token=' + this.$route.query.token
      link.target = '_blank'
      link.click()
    },
    unpublish (item) {
      this.$http.delete(Config.apiUrl + '/problems/' + item.problem_id, {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      })
        .then(() => {
          this.toast('Problem with id ' + item.problem_id + ' is unpublished.')
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
        'solution_id': item.solution_id
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
    },
    getAvailableTags () {
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) }
      }
      this.$http.get(Config.apiUrl + '/tags?mode=1', config)
        .then((response) => {
          this.available_tags = response.data
          console.log('Available Tags: ' + JSON.stringify(this.available_tags))
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    saveProblemTag ({ name, id }) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'problem_id': this.infoModal.pID,
        'tag_id': id
      }
      // console.log('Req body: ', postBody)
      this.$http.post(Config.apiUrl + '/tags/problems/', postBody, config)
        .then(data => {
          // alert('Feedback sent to student.')
          this.toast('Tag is saved.')
        })
        .catch(function (error) {
          alert(error)
        })
      // update the problem list with new tag
      var subIndex
      // console.log(this.infoModal.pID, this.message.data)
      subIndex = this.message.data.findIndex(obj => obj.problem_id === this.infoModal.pID)
      if (this.message.data[subIndex].tag === undefined) {
        this.message.data[subIndex].tag = []
      }
      this.message.data[subIndex].tag.push({'id': id, 'name': name})
    },
    remove_tag ({ name, id }) {
      // console.log('Removing: ', name, id, this.infoModal.pID)
      this.$http.delete(Config.apiUrl + '/tags/' + id + '/problems/' + this.infoModal.pID, {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      })
        .then((resp) => {
          console.log(resp)
          this.toast('Tag for problem with id ' + this.infoModal.pID + ' is removed.')
        })
        .catch(function (error) {
          // console.log(error)
          alert(error)
        })
      // update the problem list with removed tag
      var subIndex
      // console.log(this.message.data)
      subIndex = this.message.data.findIndex(obj => obj.problem_id === this.infoModal.pID)
      console.log('with Tag: ', this.message.data[subIndex])
      this.message.data[subIndex].tag = this.message.data[subIndex].tag.filter(item => item.id !== id)
    }
  },
  created: function () {
    this.getProblemList()
    this.getAvailableTags()
  }
}
</script>
<style>

.sub-action {
  margin: 5px;
}

.CodeMirror {
  height: 600px;
}

</style>

<style src="vue-multiselect/dist/vue-multiselect.min.css"></style>
