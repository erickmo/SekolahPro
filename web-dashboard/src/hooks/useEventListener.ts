import { RefObject, useEffect, useRef } from 'react'

// Overload 1: window target (element omitted or undefined)
export function useEventListener<K extends keyof WindowEventMap>(
  eventName: K,
  handler: (event: WindowEventMap[K]) => void,
  element?: undefined,
  options?: AddEventListenerOptions,
): void

// Overload 2: document target
export function useEventListener<K extends keyof DocumentEventMap>(
  eventName: K,
  handler: (event: DocumentEventMap[K]) => void,
  element: Document,
  options?: AddEventListenerOptions,
): void

// Overload 3: HTMLElement target (RefObject or direct element)
export function useEventListener<
  K extends keyof HTMLElementEventMap,
  T extends HTMLElement,
>(
  eventName: K,
  handler: (event: HTMLElementEventMap[K]) => void,
  element: RefObject<T> | T | null,
  options?: AddEventListenerOptions,
): void

// Implementation
export function useEventListener(
  eventName: string,
  handler: (event: Event) => void,
  element?: Document | RefObject<HTMLElement> | HTMLElement | null,
  options?: AddEventListenerOptions,
): void {
  const handlerRef = useRef<(event: Event) => void>(handler)

  useEffect(() => {
    handlerRef.current = handler
  }, [handler])

  useEffect(() => {
    let target: EventTarget | null = null

    if (element === undefined || element === null) {
      target = typeof window !== 'undefined' ? window : null
    } else if (element instanceof Document) {
      target = element
    } else if (element instanceof HTMLElement) {
      target = element
    } else {
      // RefObject
      target = element.current
    }

    if (target === null) return

    const listener = (event: Event) => handlerRef.current(event)

    target.addEventListener(eventName, listener, options)

    return () => {
      target!.removeEventListener(eventName, listener, options)
    }
  }, [eventName, element, options])
}
