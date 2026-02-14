<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import type { Component } from 'vue'
import {
  BookFilled as BookIcon,
  HomeFilled,
  SettingsFilled,
  DownloadFilled,
} from '@vicons/material'
import { NIcon } from 'naive-ui'
import { h, ref } from 'vue'

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}
const menuOptions: MenuOption[] = [
  {
    label: 'Chapter',
    key: 'chapter',
    icon: renderIcon(BookIcon),
  },
  {
    label: 'Manga Alternative',
    key: 'manga-alternative',
    disabled: true,
    icon: renderIcon(BookIcon),
  },
]

const router = useRouter()
const route = useRoute()
const headerKey = ref('home')
const headerMenu = [
  {
    label: 'Home',
    key: 'home',
    icon: renderIcon(HomeFilled),
  },
  {
    label: 'Download',
    key: 'download',
    icon: renderIcon(DownloadFilled),
  },
  {
    label: 'Settings',
    key: 'settings',
    icon: renderIcon(SettingsFilled),
  },
]

const handleHeaderMenu = (key: string) => {
  if (key === 'home') {
    router.push('/')
  } else if (key === 'settings') {
    router.push('/settings')
  } else if (key === 'download') {
    router.push('/download')
  }
}

const activeKey = ref<string | null>(null)
const collapsed = ref(true)
</script>

<template>
  <div>
    <n-config-provider :theme="darkTheme">
      <n-global-style />
      <n-message-provider>
        <n-dialog-provider>
          <n-layout>
            <n-space vertical>
              <n-layout has-sider>
                <n-layout-sider
                  bordered
                  collapse-mode="width"
                  :collapsed-width="0"
                  :width="240"
                  :collapsed="collapsed"
                  show-trigger
                  @collapse="collapsed = true"
                  @expand="collapsed = false"
                >
                  <n-menu
                    v-model:value="activeKey"
                    :collapsed="collapsed"
                    :collapsed-width="64"
                    :collapsed-icon-size="22"
                    :options="menuOptions"
                  />
                </n-layout-sider>
                <n-layout>
                  <n-layout-header bordered>
                    <n-menu
                      mode="horizontal"
                      v-model:value="headerKey"
                      :options="headerMenu"
                      @update:value="handleHeaderMenu"
                    />
                  </n-layout-header>
                  <n-layout-content
                    :content-style="`padding: ${route.name === 'read' ? '0' : '24px'}; height: calc(100vh - 66px);`"
                  >
                    <div id="main">
                      <router-view></router-view>
                    </div>
                    <div id="reader-fs"></div>
                  </n-layout-content>
                  <n-layout-footer bordered>
                    <div class="px-5">Mangav5 - All in one Manga manager</div>
                  </n-layout-footer>
                </n-layout>
              </n-layout>
            </n-space>
          </n-layout>
        </n-dialog-provider>
      </n-message-provider>
    </n-config-provider>
  </div>
</template>

<style scoped></style>
<style>
body {
  background: black;
}
:root {
  color-scheme: light dark;
}
</style>
