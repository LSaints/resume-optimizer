import type { RenderResponse } from '../types/render'
import { get } from './api'

const BASE_URL = 'http://localhost:8080/v1'

export async function getRenderSVG(optimizationID: string): Promise<string> {
  const res = await get<RenderResponse>(`/optimizations/${optimizationID}/render`)
  return res.svgContent
}

export function getDownloadPDFURL(optimizationID: string): string {
  return `${BASE_URL}/optimizations/${optimizationID}/render/pdf`
}
