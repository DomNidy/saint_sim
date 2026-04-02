import { Button } from '#/components/ui/button'
import { authClient } from '#/lib/auth-client'
import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/auth/sign-in/')({
  component: RouteComponent,
})

function RouteComponent() {

  const signInWithDiscord = async () => {
    const data = await authClient.signIn.social({ provider: "discord" })
    console.log(`res data: ${data}`)
  }
  return <div>
    <Button onClick={signInWithDiscord}>Sign in w/ Discord</Button>
  </div>
}
