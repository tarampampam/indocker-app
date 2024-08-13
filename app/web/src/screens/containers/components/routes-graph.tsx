import React from 'react'

export type GraphPoints = Map<string, { isBuiltIn?: boolean; url: URL | null }>

export default function RoutesGraph({
  loading,
  points,
}: {
  loading?: boolean
  points?: Readonly<GraphPoints> | null
}): React.JSX.Element {
  console.log(loading, points)

  // https://reactflow.dev/learn

  return (
    <div style={{ width: '100%', height: '100px' }}>{/*<ReactFlow nodes={initialNodes} edges={initialEdges} />*/}</div>
  )
}
