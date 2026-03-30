import * as React from 'react'

import { cn } from '#/lib/utils'

const Input = React.forwardRef<HTMLInputElement, React.ComponentProps<'input'>>(
  ({ className, type, ...props }, ref) => {
    return (
      <input
        ref={ref}
        type={type}
        className={cn(
          'flex h-11 w-full rounded-2xl border border-[var(--line)] bg-white/70 px-4 py-2 text-sm text-[var(--sea-ink)] shadow-[0_1px_0_var(--inset-glint)_inset] transition placeholder:text-[var(--sea-ink-soft)]/70 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[var(--lagoon)] focus-visible:ring-offset-2 focus-visible:ring-offset-transparent disabled:cursor-not-allowed disabled:opacity-50 dark:bg-[rgba(13,28,32,0.85)]',
          className,
        )}
        {...props}
      />
    )
  },
)

Input.displayName = 'Input'

export { Input }
