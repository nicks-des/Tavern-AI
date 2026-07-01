const MOCK_REPLIES: Record<string, string[]> = {
  alice: [
    'That is a wonderful question! Let me think about it for a moment...\n\nFrom a scientific perspective, what you are describing touches on several fascinating fields — cognitive science, information theory, and even a bit of philosophy of mind.\n\nI would say the most intriguing part is how these ideas connect. Each piece of the puzzle reveals something new about the whole picture.',
    'You know, I have been pondering something similar lately.\n\nThe more I read about emergent properties in complex systems, the more I realize that simple rules can create incredibly rich behavior.\n\nWhat do you think drives that emergence? Is it the rules themselves, or the interactions between components?',
    'I love that we are having this conversation! It reminds me of something Carl Sagan once said — that understanding is a kind of ecstasy.\n\nCan I ask you a follow-up? What sparked your interest in this topic originally?',
  ],
  ryu: [
    'Hmph. Words are wind. Action reveals truth.\n\nIf you wish to understand the way of the warrior, you must first understand yourself. The blade is merely an extension of the spirit.',
    'I have traveled across many lands, and I have learned this: strength without purpose is destruction. Purpose without strength is empty.\n\nWhich do you seek?',
    '...You show promise. But promise is not mastery.\n\nTrain. Every day. When you think you have trained enough, train more.',
  ],
  luna: [
    'Oh, what a lovely thought! It shimmers like morning light on the water...\n\nI think the world needs more people who stop to notice the little things — the way the clouds change shape, the sound of rain on the roof, the smell of salt in the air.\n\nDo you take time to notice those things?',
    'You know what I have learned from painting the same sea for so many years?\n\nNo two waves are the same. Not really. Each one carries its own story, its own dance, its own little moment of being alive.\n\nMaybe people are like that too.',
    'The lighthouse keeper before me used to say that the light does not just guide ships — it guides lost souls home.\n\nI am not sure if I believe that literally, but I do think we all need a light sometimes.\n\nWhat guides you when you feel lost?',
  ],
}

const FALLBACK = [
  'That is an interesting perspective. Tell me more about what you think.',
  'I see. And how does that make you feel?',
  'Let me ponder this carefully... I think there is more to explore here.',
]

export function generateMockReply(characterId: string, userMessage: string): string {
  const replies = MOCK_REPLIES[characterId]
  if (replies) {
    return replies[Math.floor(Math.random() * replies.length)]
  }
  return FALLBACK[Math.floor(Math.random() * FALLBACK.length)]
}

export function simulateStreaming(
  fullText: string,
  onToken: (token: string) => void,
  onDone: () => void,
  speed = 30
) {
  let i = 0
  const interval = setInterval(() => {
    if (i < fullText.length) {
      const chunk = fullText.slice(i, i + 3)
      onToken(chunk)
      i += 3
    } else {
      clearInterval(interval)
      onDone()
    }
  }, speed)
  return () => clearInterval(interval)
}
