import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/_protected/account')({
  component: RouteComponent,
})

function RouteComponent() {
  const { session } = Route.useRouteContext()
  return <div>Hello {session.user.name}, {session.session.id}</div>
}
