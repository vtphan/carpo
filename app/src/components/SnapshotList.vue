<template>
    <div>
      <b-card no-body>
        <b-tabs card>
          <b-tab active>
            <template #title>
              <div v-on:click="getSnapshotList()">  <a v-if="message.data">({{ message.data.length}})</a> </div>
            </template>
              <!-- <div style="float:right; position: absolute; top: 6px; left: calc(100% - 165px);">
                <b-dropdown no-caret>
                  <template #button-content>
                    <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Order By
                  </template>
                  <b-dropdown-item href="#" @click="setSorting('creation_time')">LastActive At</b-dropdown-item>
                  <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
                </b-dropdown>
              </div> -->
              <div v-if="isLoading">
                <div>LOADING...</div>
              </div>
              <b-card-text v-else>
                <div>
                <v-row class="five-cols">
                <!-- <div class="items" > -->
                    <b-card
                      class="item"
                      v-b-modal = "'myModal2'"
                      :style="{'border-color': setborderColor(items.created_at)}"
                      v-for="items in message.data" :key="items.id"
                      @click="sendInfo(items)">
                        <template #header>
                          <div class="box-header d-flex justify-content-between align-items-center">
                            {{ items.id }} : {{ items.problem_id }}
                            <b-icon v-if="items.on_watch" icon="flag-fill" scale="2"></b-icon>
                          </div>
                        </template>
                        <b-card-text >
                           {{ items.student_name }}
                        </b-card-text>
                        <template #footer>
                          <small>
                            Active {{ timeDiff(items.created_at) }} ago
                          </small>
                        </template>
                      </b-card>
                  <!-- </div> -->
                </v-row>
              </div>
              <b-modal id="myModal2" size="xl" :hide-footer="true">
                <template #modal-title>
                  <div class="box-header d-flex justify-content-between align-items-center">
                    <div style="margin-right: 20px;"> Snapshot {{ timeDiff(selectedSub.created_at) }} ago </div>
                    <b-icon v-if="selectedSub.on_watch" icon="flag-fill" scale="2"></b-icon>
                  </div>
                </template>
                <codemirror v-model="selectedSub.code" :options="cmOptions" />
                <b-row>
                  <b-col cols="6" >
                    <div style="text-align: left">
                      <div v-if="selectedSub.on_watch" class="row">
                          <b-button class="btn-secondary" @click="unwatchSub(selectedSub);">Unwatch</b-button>
                      </div>
                      <div v-else class="row">
                        <b-button class="btn-secondary" @click="watchSubmission(selectedSub);">Watch</b-button>
                        <b-form-input style="width: 60%; height: auto;" v-model="reason" placeholder="Reason to set on Watch (Optional)"></b-form-input>
                      </div>
                    </div>
                  </b-col>
                  <b-col cols="6" >
                    <div style="text-align: right">
                      <b-button-group>
                        <b-button class="btn-secondary" @click="sendFeedback(selectedSub, selectedSub.id)">Send Feedback</b-button>
                      </b-button-group>
                    </div>
                  </b-col>
                </b-row>
              </b-modal>
            </b-card-text>
          </b-tab>
        </b-tabs>
      </b-card>
    </div>
</template>

<script>
import { codemirror } from 'vue-codemirror'
import 'codemirror/lib/codemirror.css'
// language
import 'codemirror/mode/python/python.js'

// theme css
import 'codemirror/theme/duotone-light.css'

import * as Config from '../config'
import moment from 'moment'

export default {
  name: 'SnapshotList',
  components: {
    codemirror
  },
  data: () => ({
    message: '',
    selectedSub: '',
    watchSubs: '',
    reason: '',
    sorting: 'creation_time',
    watchedSub: '',
    isLoading: true,
    cmOptions: {
      autoRefresh: true,
      tabSize: 4,
      styleActiveLine: true,
      lineNumbers: true,
      line: true,
      mode: 'application/x-httpd-python',
      lineWrapping: true,
      theme: 'duotone-light'
    }
  }),
  methods: {
    sendInfo (item) {
      // console.log('SendInfo:', item)
      this.selectedSub = item
      this.reason = ''
    },
    getImagePath () {
      return require('../assets/code-block-1.png')
    },
    timeDiff (dbTimestamp) {
      return moment.duration(moment().diff(moment(dbTimestamp))).humanize()
      // https://stackoverflow.com/questions/18623783/get-the-time-difference-between-two-datetimes
    },
    getColor (value) {
      // value from 0 to 1
      var hue = ((1 - value) * 120).toString(10)
      return ['hsl(', hue, ',100%,50%)'].join('')
    },
    setborderColor (dbTimestamp) {
      const limit = 5
      var ago = moment.duration(moment().diff(moment(dbTimestamp))).minutes()
      if (ago > 5) {
        ago = 5
      }
      var value = ago / limit
      return this.getColor(value)
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
    watchSubmission (sub) {
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token }
      }
      let postBody = {
        'student_id': sub.student_id,
        'submission_id': sub.id,
        'problem_id': sub.problem_id,
        'reason': this.reason,
        'mode': 1
      }

      this.$http.post(Config.apiUrl + '/snapshots/watch', postBody, config)
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is on watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is on watch list.')
          this.message.data = this.message.data.map(obj => {
            if (obj.id === sub.id) {
              return { ...obj, on_watch: 1 }
            }
            return obj
          })
          this.selectedSub.on_watch = 1
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    unwatchSub (sub) {
      console.log(sub)
      this.$http.delete(Config.apiUrl + '/snapshots/watch', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {id: sub.watch_id}
      })
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is removed from the watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is removed from the watch list.')
          this.message.data = this.message.data.map(obj => {
            if (obj.id === sub.id) {
              return { ...obj, on_watch: 0 }
            }
            return obj
          })
          this.selectedSub.on_watch = 0
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

      this.$http.post(Config.apiUrl + '/submissions/grades', postBody, config)
        .then(data => {
          // alert('Feedback sent to student.')
          this.toast('Feedback is sent to student.')
        })
    },
    getSnapshotList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        params: {'sort_by': this.sorting}
      }
      this.$http.get(Config.apiUrl + '/snapshots/teachers', config)
        .then((response) => {
          console.log('Snapshot: ', response)
          this.message = response.data
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    getWatchedSubsList: function () {
      this.isLoading = true
      const config = {
        headers: { Authorization: 'Bearer '.concat(this.$route.query.token) },
        params: {'sort_by': this.sorting}
      }
      this.$http.get(Config.apiUrl + '/snapshots/watch', config)
        .then((response) => {
          this.watchSubs = response.data
          // console.log('watched' + JSON.stringify(this.watchSubs))
          this.isLoading = false
        })
        .catch((error) => {
          console.log('Error', error)
          this.toast('Unauthorized Access.')
        })
    },
    setSorting (params) {
      this.sorting = params
      this.getSnapshotList()
      this.getWatchedSubsList()
    }
  },
  created: function () {
    this.getSnapshotList()
    this.getWatchedSubsList()
    // setInterval(() => this.getSnapshotList(), 10000)
  }
}
</script>
<style>

.five-cols {
  display: grid;
  grid-template-columns: repeat(10, 1fr);
  column-gap: 8px;
  background-color: rgb(206, 209, 212);
  padding: 5px;
  /* text-align: left; */
}

.item .card-header {
  padding: 0.25rem 0.25rem;
  font-weight: 300;
  font-size: 12px;
}

.card-body {
  padding: 2px;
  height: 55px;
}

b-card-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.card-footer {
  padding: 0.25rem 0.25rem;
}

small, .small {
    font-size: 11px;
    font-weight: 400;
}

br {
  content: "";
  margin: -2em;
  display: block;
  font-size: 24%;
}

.item {
  background-color: rgb(235, 229, 229);
  display: inline-block;
  width: 100%;
  max-width: 14rem;
  margin: 5px;
  font-size: 12px;
  font-weight: 300;
  border-width: 3px;
}

.tab-content .active {
    padding: 0px;
}

button {
  margin: 5px
}

input:placeholder-shown {
   font-style: italic;
}

.CodeMirror {
  height: 600px;
}

.box-header {
  margin: 5px;
  padding-left: 10px;
  padding-right: 10px;
}
</style>
