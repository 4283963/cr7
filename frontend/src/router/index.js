import { createRouter, createWebHashHistory } from 'vue-router'
import ScriptList from '@/views/ScriptList.vue'
import Editor from '@/views/Editor.vue'

const routes = [
  { path: '/', name: 'home', component: ScriptList },
  { path: '/editor/:id', name: 'editor', component: Editor, props: true }
]

const router = createRouter({
  history: createWebHashHistory(),
  routes
})

export default router
