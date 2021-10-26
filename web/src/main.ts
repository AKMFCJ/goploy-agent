import { createApp } from 'vue'
import 'normalize.css/normalize.css' // A modern alternative to CSS resets
import '@/styles/index.scss' // global css
import router from '@/router'
import i18n from '@/lang' // internationalization
import SvgIcon from '@/components/SvgIcon/index.vue' // svg component
import ElementPlus from 'element-plus'
import 'element-plus/packages/theme-chalk/src/index.scss'
import 'vite-plugin-svg-icons/register' // register svg sprite map
import App from './App.vue'
const app = createApp(App)
app.use(ElementPlus, {
  size: 'mini',
})
app.use(router)
app.use(i18n)
app.component('SvgIcon', SvgIcon)

app.mount('#app')
