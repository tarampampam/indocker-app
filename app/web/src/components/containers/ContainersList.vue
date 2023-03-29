<template>
  <n-list show-divider bordered hoverable clickable style="background-color: transparent">
    <n-list-item
      v-for="container in containers"
      :key="container.id"
      v-on:click="goto(useRouter(), RouteName.ContainerLogs, { id: container.id })"
      :class="{ active: id() === container.id }"
    >
      <template #prefix>
        <n-icon-wrapper
          :size="14"
          :border-radius="14"
          :color="theme.defaultColor"
        /><!-- change the color here -->
      </template>
      <template #default>
        <n-thing>
          <template #header>
            <div class="container-name">
              <span class="wrap" :title="container.name.length > 10 ? container.name : undefined">
                {{ container.name }}
              </span>
            </div>
          </template>
          <template #description>
            <n-space size="small" justify="start" class="container-tags">
              <n-tag v-for="tag in container.tags" :key="tag" type="info" size="small" round>
                <div class="tag">
                  <span class="wrap" :title="tag.length > 7 ? tag : undefined">
                    {{ tag }}
                  </span>
                </div>
              </n-tag>
            </n-space>
          </template>
        </n-thing>
      </template>
      <template #suffix>
        <n-icon :size="25" :depth="4">
          <arrow-icon />
        </n-icon>
      </template>
    </n-list-item>
  </n-list>
</template>

<script lang="ts">
import { defineComponent } from 'vue'
import { NIcon, NIconWrapper, NList, NListItem, NSpace, NTag, NThing, useThemeVars } from 'naive-ui'
import { ArrowForwardCircleOutline as ArrowIcon } from '@vicons/ionicons5'
import type { Container } from '@/components/containers/ViewAll.vue'
import { RouteName, current, goto, id } from '@/router'
import { useRouter } from 'vue-router';

export default defineComponent({
  methods: {useRouter},
  props: {
    containers: {
      type: Array as () => Container[],
      required: true
    }
  },
  setup() {
    return {
      RouteName,
      current,
      goto,
      id
    }
  },
  components: {
    NIconWrapper,
    NListItem,
    NThing,
    NSpace,
    NList,
    NIcon,
    NTag,
    ArrowIcon
  },
  data(): {
    theme: {
      activeColor: string
      successColor: string
      warningColor: string
      errorColor: string
      defaultColor: string
    }
  } {
    const theme = useThemeVars()

    return {
      theme: {
        activeColor: theme.value.infoColor,
        successColor: theme.value.successColor,
        warningColor: theme.value.warningColor,
        errorColor: theme.value.errorColor,
        defaultColor: theme.value.iconColorDisabled
      }
    }
  }
})
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
