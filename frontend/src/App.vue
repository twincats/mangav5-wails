<script setup lang="ts">
import { darkTheme } from 'naive-ui'
import type { MenuOption } from 'naive-ui'
import type { Component } from 'vue'
import {
  BookFilled as BookIcon,
  PersonAddAltRound as PersonIcon,
  WineBarRound as WineIcon,
} from '@vicons/material'
import { NIcon } from 'naive-ui'
import { h, ref } from 'vue'

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}
const menuOptions: MenuOption[] = [
  {
    label: 'Hear the Wind Sing',
    key: 'hear-the-wind-sing',
    icon: renderIcon(BookIcon),
  },
  {
    label: 'Pinball 1973',
    key: 'pinball-1973',
    icon: renderIcon(BookIcon),
    disabled: true,
    children: [
      {
        label: 'Rat',
        key: 'rat',
      },
    ],
  },
  {
    label: 'A Wild Sheep Chase',
    key: 'a-wild-sheep-chase',
    disabled: true,
    icon: renderIcon(BookIcon),
  },
  {
    label: 'Dance Dance Dance',
    key: 'Dance Dance Dance',
    icon: renderIcon(BookIcon),
    children: [
      {
        type: 'group',
        label: 'People',
        key: 'people',
        children: [
          {
            label: 'Narrator',
            key: 'narrator',
            icon: renderIcon(PersonIcon),
          },
          {
            label: 'Sheep Man',
            key: 'sheep-man',
            icon: renderIcon(PersonIcon),
          },
        ],
      },
      {
        label: 'Beverage',
        key: 'beverage',
        icon: renderIcon(WineIcon),
        children: [
          {
            label: 'Whisky',
            key: 'whisky',
          },
        ],
      },
      {
        label: 'Food',
        key: 'food',
        children: [
          {
            label: 'Sandwich',
            key: 'sandwich',
          },
        ],
      },
      {
        label: 'The past increases. The future recedes.',
        key: 'the-past-increases-the-future-recedes',
      },
    ],
  },
]

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
                  :collapsed-width="64"
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
                  <n-layout-header bordered> Yiheyuan Road </n-layout-header>
                  <n-layout-content
                    content-style="padding: 24px; height: calc(100vh - 50px);"
                  >
                    <router-view></router-view>
                  </n-layout-content>
                  <n-layout-footer bordered> Chengfu Road </n-layout-footer>
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
