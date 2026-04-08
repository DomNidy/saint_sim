import { getSession } from '@/lib/auth.functions'
import { createFileRoute, Outlet, redirect } from '@tanstack/react-router'

/**
 * Tanstack pathless layout route that redirects if accessed
 * without a session. 
 * 
 * Every child route of this route requires authentication
 * to access.
 */
export const Route = createFileRoute('/_protected')({
  beforeLoad: async ({ location }) => {
    const session = await getSession()

    if (!session) {
      throw redirect({
        /**
         * Route to redirect to when not authed
         */
        to: "/auth/sign-in",
        search: {
          redirect: location.href
        }
      })
    }

    // add the session into context
    return { session: session }
  },
  component: () => <Outlet />
})
