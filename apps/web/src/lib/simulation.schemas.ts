import { z } from 'zod'

export const simulationRegions = ['us', 'eu', 'kr', 'tw', 'cn'] as const

export const simulationRealms = [
  'thrall',
  'hydraxis',
  'silvermoon',
  'draenor',
] as const

export const simulationRequestSchema = z.object({
  region: z.enum(simulationRegions),
  realm: z.enum(simulationRealms),
  character_name: z.string().trim().min(1, 'Character name is required'),
})

export type SimulationRequestInput = z.infer<typeof simulationRequestSchema>

export const simulationResultLookupSchema = z.object({
  requestId: z.string().uuid(),
})

export type SimulationResultLookupInput = z.infer<
  typeof simulationResultLookupSchema
>
