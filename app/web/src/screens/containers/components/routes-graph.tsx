import type {
  GraphCategoryItemOption,
  GraphEdgeItemOption,
  GraphNodeItemOption,
} from 'echarts/types/src/chart/graph/GraphSeries.d.ts'
import React, { useEffect, useState } from 'react'
import { Chart } from '~/shared/components'

export type GraphPoints = Map<string, { isBuiltIn?: boolean; url: URL | null }>

type ChartData = {
  nodes: Array<GraphNodeItemOption>
  links: Array<GraphEdgeItemOption>
  categories: Array<GraphCategoryItemOption>
}

enum NodeCategories {
  Empty = 0,
  User = 1,
  Service = 3,
}

export default function RoutesGraph({
  loading,
  points,
}: {
  loading?: boolean
  points?: Readonly<GraphPoints> | null
}): React.JSX.Element {
  const [data, setData] = useState<ChartData | null>(null)

  useEffect(() => {
    if (!points) {
      return
    }

    // set initial data
    const newData: ChartData = {
      nodes: [],
      links: [],
      categories: Object.values(NodeCategories).map((category) => ({ name: category.toString() })),
    }

    // add nodes and links for each route
    for (const [hostname, entry] of points.entries()) {
      if (hostname === '') {
        continue
      }

      // 'foo.bar.indocker.app' -> ['foo.bar.indocker.app', 'bar.indocker.app', 'indocker.app', 'app']
      const subdomains = hostname.split('.').map((_, i, arr) => arr.slice(i).join('.'))

      subdomains.pop() // remove the last element - the root domain

      for (let i = 0; i < subdomains.length; i++) {
        const parent: string | null = i < subdomains.length - 1 ? subdomains[i + 1] : null
        const current: string = subdomains[i]

        if (newData.nodes.some((node) => node.id === current)) {
          continue
        }

        const hasUrls = points.has(current) && !!points.get(current)?.url

        newData.nodes.push({
          id: current,
          name: current,
          symbol: entry.isBuiltIn ? 'roundRect' : 'circle',
          symbolSize: entry.isBuiltIn ? 50 : hasUrls ? 30 : 20,
          category: ((): NodeCategories => {
            switch (true) {
              case entry.isBuiltIn:
                return NodeCategories.Service
              case hasUrls:
                return NodeCategories.User
              default:
                return NodeCategories.Empty
            }
          })(),
          label: {
            fontFamily: 'monospace',
            borderType: 'solid',
            borderWidth: 1,
            fontSize: (hasUrls ? 1.35 : 1) + 'em',
            silent: !hasUrls,
          },
          cursor: hasUrls ? 'pointer' : 'default',
          emphasis: { focus: 'adjacency' },
        })

        if (parent) {
          newData.links.push({
            source: current,
            target: parent,
            lineStyle: { width: 5, type: hasUrls ? 'solid' : 'dotted' },
          })
        }
      }
    }

    setData(newData)
  }, [points])

  return (
    <Chart
      option={{
        tooltip: { show: false },
        series: [
          {
            name: 'Containers network', // Series name used for displaying in tooltip and filtering with legend
            type: 'graph', // https://echarts.apache.org/en/option.html#series-graph
            layout: 'force', // https://echarts.apache.org/en/option.html#series-graph.layout
            data: data?.nodes,
            links: data?.links,
            categories: data?.categories,
            roam: false, // Whether to enable mouse zooming and translating
            label: { show: true, position: 'bottom', distance: 15, formatter: '{b}' },
            lineStyle: { color: 'source', curveness: 0.1, opacity: 0.4 },
            force: { initLayout: 'circular', repulsion: 500, gravity: 0.1, edgeLength: 150, friction: 0.1 },
          },
        ],
      }}
      loading={loading}
      style={{ height: 1300 }}
      events={{
        click: (e) => {
          if (e.dataType === 'node' && points?.has(e.name)) {
            const entry = points.get(e.name)

            if (entry?.url) {
              const a = document.createElement('a')

              a.href = entry.url.toString()
              a.target = '_blank'
              a.rel = 'noopener noreferrer'
              a.click()
              a.remove() // is this necessary?
            }
          }
        },
      }}
    />
  )
}
