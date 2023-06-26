<template>
    <div>
      <b-card no-body>
        <b-tabs card>
          <b-tab active>
            <template #title>
              <div v-on:click="getSnapshotList()"> Snapshot <a v-if="message.data">({{ message.data.length}})</a> </div>
            </template>
              <div style="float:right; position: absolute; top: 6px; left: calc(100% - 165px);">
                <b-dropdown no-caret>
                  <template #button-content>
                    <b-icon icon="gear-fill" aria-hidden="true"></b-icon> Order By
                  </template>
                  <b-dropdown-item href="#" @click="setSorting('creation_time')">LastActive At</b-dropdown-item>
                  <b-dropdown-item href="#" @click="setSorting('name')">Name</b-dropdown-item>
                </b-dropdown>
              </div>
              <div  v-if="isLoading">
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
                            SUBID: {{ items.id }}
                            <br>
                            PID: {{ items.problem_id }}
                        </template>
                        <b-card-text >
                           {{ items.student_name }}
                        </b-card-text>
                        <template #footer>
                          <small>
                            Last Active: {{ timeDiff(items.created_at) }} ago
                          </small>
                        </template>
                      </b-card>
                  <!-- </div> -->
                </v-row>
              </div>
              <b-modal id="myModal2" title="Snapshot View" size="xl" ok-only ok-variant="secondary" ok-title="Send Feedback" @ok="sendFeedback(selectedSub, selectedSub.id)">
                <codemirror v-model="selectedSub.code" :options="cmOptions" />
                  <div style="float:right; position: absolute; bottom: -55px; right: calc(100% - 85px);">
                    <b-button class="btn-secondary" @click="watchSubmission(selectedSub); $bvModal.hide('myModal2')">Watch</b-button>
                  </div>
              </b-modal>
            </b-card-text>
          </b-tab>
          <b-tab >
            <template #title>
              <div v-on:click="getWatchedSubsList()"> Watched <a v-if="watchSubs.data">({{ watchSubs.data.length}})</a>
              </div>
            </template>
            <div v-if="isLoading">
              <p>LOADING...</p>
            </div>
            <b-card-text v-else>
              <div>
                <!-- <div class="items" > -->
                <v-row class="five-cols">
                    <b-card
                      class="item"
                      v-b-modal = "'watchModal'"
                      :style="{'border-color': setborderColor(items.created_at)}"
                      v-for="items in watchSubs.data" :key="items.id"
                      @click="sendInfo(items)">
                        <template #header >
                          SUBID: {{ items.submission_id }}
                          <br>
                          PID: {{ items.problem_id }}
                        </template>
                        <b-card-text >
                            From: {{ items.student_name }}
                        </b-card-text>
                        <template #footer>
                          <small>
                            Last Active: {{ timeDiff(items.created_at) }} ago
                          </small>
                        </template>
                    </b-card>
                <!-- </div> -->
                </v-row>
              </div>
              <b-modal id="watchModal" title="On Watch Snapshot" size="xl" ok-only ok-variant="secondary" ok-title="Send Feedback" @ok="sendFeedback(selectedSub, selectedSub.submission_id)">
                  <codemirror v-model="selectedSub.code" :options="cmOptions" />
                  <div style="float:right; position: absolute; bottom: -55px; right: calc(100% - 95px);">
                    <b-button class="btn-secondary" @click="unwatchSub(selectedSub); $bvModal.hide('watchModal')">Unwatch</b-button>
                  </div>
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
      console.log('SendInfo:', item)
      this.selectedSub = item
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
        'problem_id': sub.problem_id
      }

      this.$http.post(Config.apiUrl + '/snapshots/watch', postBody, config)
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is on watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is on watch list.')
        })
        .catch(function (error) {
          console.log(error)
          // alert(error)
        })
    },
    unwatchSub (sub) {
      this.$http.delete(Config.apiUrl + '/snapshots/watch', {
        headers: { Authorization: 'Bearer ' + this.$route.query.token },
        data: {watch_id: sub.id}
      })
        .then(() => {
          // alert('Snapshot  with id ' + sub.student_id + ' is removed from the watch list.')
          this.toast('Snapshot of student with id ' + sub.student_id + ' is removed from the watch list.')
          this.watchSubs.data = this.watchSubs.data.filter(item => item.id !== sub.id)
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
        'teacher_id': this.$route.query.id,
        'code': submission.code
      }

      this.$http.post(Config.apiUrl + '/teachers/feedbacks', postBody, config)
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
      this.$http.get(Config.apiUrl + '/teachers/snapshots', config)
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
/* Make it responsive */
@media only screen and (max-width: 1000px) {
  .items {
    column-count: 4;
  }
}

@media only screen and (max-width: 600px) {
  .items {
    column-count: 3;
  }
}

@media only screen and (max-width: 400px) {
  .items {
    column-count: 2;
  }
}

@media only screen and (max-width: 100px) {
  .items {
    column-count: 1;
  }
}

</style>
