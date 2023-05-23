import Vue from 'vue'
import Router from 'vue-router'
import ProblemList from '@/components/ProblemList.vue'
import SubmissionList from '@/components/SubmissionList.vue'
import SnapshotList from '@/components/SnapshotList.vue'

Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      redirect: '/problems',
      name: 'Home'

    },
    {
      path: '/problems',
      name: 'Problems',
      component: ProblemList

    },
    {
      path: '/submissions',
      name: 'Submissions',
      component: SubmissionList

    },
    {
      path: '/sanpshots',
      name: 'Snapshots',
      component: SnapshotList

    }
  ]
})
