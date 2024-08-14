import React, { type ReactNode } from 'react'
import { motion } from 'framer-motion'

export default function Component({ children }: { children: ReactNode }): React.JSX.Element {
  return (
    <motion.div
      initial="hidden"
      animate="enter"
      exit="exit"
      variants={{
        hidden: { opacity: 0, filter: 'blur(5px)' },
        enter: { opacity: 1, filter: 'blur(0)' },
        exit: { opacity: 0, filter: 'blur(5px)' },
      }}
      transition={{ duration: 0.1, type: 'easeInOut' }}
      style={{ height: '100%', position: 'relative' }}
    >
      {children}
    </motion.div>
  )
}
