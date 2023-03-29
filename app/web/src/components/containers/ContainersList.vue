<template>
  <NList show-divider bordered hoverable clickable style="background-color: transparent">
    <NListItem
      v-for="container in containers"
      :key="container.id"
      :class="{ active: router.currentRoute.value.params.id === container.id }"
      @click="router.push({ name: RouteName.ContainerLogs, params: { id: container.id } })"
    >
      <template #prefix>
        <NIconWrapper
          :size="14"
          :border-radius="14"
          :color="theme.defaultColor"
        /><!-- change the color here -->
      </template>
      <template #default>
        <NThing>
          <template #header>
            <div class="container-name">
              <span class="wrap" :title="container.name.length > 10 ? container.name : undefined">
                {{ container.name }}
              </span>
            </div>
          </template>
          <template #description>
            <NSpace size="small" justify="start" class="container-tags">
              <NTag v-for="tag in container.tags" :key="tag" type="info" size="small" round>
                <div class="tag">
                  <span class="wrap" :title="tag.length > 7 ? tag : undefined">
                    {{ tag }}
                  </span>
                </div>
              </NTag>
            </NSpace>
          </template>
        </NThing>
      </template>
      <template #suffix>
        <NIcon :size="25" :depth="4">
          <arrow-icon />
        </NIcon>
      </template>
    </NListItem>
  </NList>
</template>

<script setup lang="ts">
import { defineProps } from 'vue'
import { NIcon, NIconWrapper, NList, NListItem, NSpace, NTag, NThing, useThemeVars } from 'naive-ui'
import { ArrowForwardCircleOutline as ArrowIcon } from '@vicons/ionicons5'
import type { Container } from './types'
import { RouteName } from '@/router'
import { useRouter } from 'vue-router'

defineProps({
  containers: {
    type: Array as () => Container[],
    required: true
  }
})

const router = useRouter()
const themeVars = useThemeVars()
const theme = {
  activeColor: themeVars.value.infoColor,
  successColor: themeVars.value.successColor,
  warningColor: themeVars.value.warningColor,
  errorColor: themeVars.value.errorColor,
  defaultColor: themeVars.value.iconColorDisabled
}
</script>

<style lang="scss" scoped>
$width: 150px;

.container-name {
  width: $width;
  display: inline-flex;
  box-sizing: border-box;
  padding-left: 0.4em;
  font-size: 1.2em;

  .wrap {
    text-overflow: ellipsis;
    white-space: nowrap;
    overflow: hidden;
  }
}

.container-tags {
  width: $width;

  .tag {
    display: inline-flex;
    max-width: $width - 15px; // 15px for "..."

    .wrap {
      text-overflow: ellipsis;
      white-space: nowrap;
      overflow: hidden;
    }
  }
}

.active {
  //box-shadow: 0 -2px v-bind('theme.activeColor') inset;

  &:after {
    content: '';
    position: absolute;
    width: 3px;
    height: 100%;
    left: 0;
    background: linear-gradient(to top, #12c2e9, #c471ed, #f64f59);
  }
}
</style>
