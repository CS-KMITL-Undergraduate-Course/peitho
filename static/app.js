const btn = document.getElementById('btnConsult');
const inputText = document.getElementById('inputText');
const contextSelect = document.getElementById('contextSelect');
const resultPanel = document.getElementById('resultPanel');

btn.addEventListener('click', async () => {
  const text = inputText.value.trim();
  if (!text) return alert('Please provide an offering (text).');

  // UI Loading state
  btn.disabled = true;
  btn.textContent = 'Consulting...';
  resultPanel.style.opacity = '1';
  resultPanel.style.pointerEvents = 'auto';

  const warningEl = document.getElementById('warningLevel');
  warningEl.textContent = 'ANALYZING...';
  warningEl.style.color = '';
  warningEl.style.borderColor = '';

  document.getElementById('reasoningText').textContent = 'The Oracle is peering into the threads of fate...';

  try {
    const response = await fetch('/api/moderate', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: text, context: contextSelect.value })
    });

    if (!response.ok) throw new Error('The Oracle is silent (Server Error).');

    const data = await response.json();

    // Update UI Scores
    updateScore('profanity', data.scores.profanity);
    updateScore('privacy', data.scores.privacy);
    updateScore('phishing', data.scores.phishing);
    updateScore('violence', data.scores.violence_selfharm);

    // Determine Overall Warning
    const maxScore = Math.max(
      data.scores.profanity,
      data.scores.privacy,
      data.scores.phishing,
      data.scores.violence_selfharm
    );

    if (maxScore > 0.8) {
      warningEl.textContent = 'MORTAL WARNING';
      warningEl.style.color = '#d32f2f';
      warningEl.style.borderColor = '#d32f2f';
    } else if (maxScore > 0.4) {
      warningEl.textContent = 'CAUTION ADVISED';
      warningEl.style.color = '#f57c00';
      warningEl.style.borderColor = '#f57c00';
    } else {
      warningEl.textContent = 'SAFE PASSAGE';
      warningEl.style.color = '#388e3c';
      warningEl.style.borderColor = '#388e3c';
    }

    // Update Reasoning
    document.getElementById('reasoningText').textContent = data.reasoning;

  } catch (error) {
    warningEl.textContent = 'ERROR';
    document.getElementById('reasoningText').textContent = error.message;
  } finally {
    btn.disabled = false;
    btn.textContent = 'Consult Peitho';
  }
});

function updateScore(id, score) {
  const percentage = Math.round(score * 100);
  document.getElementById(`val-${id}`).textContent = `${percentage}%`;
  document.getElementById(`bar-${id}`).style.width = `${percentage}%`;
}
