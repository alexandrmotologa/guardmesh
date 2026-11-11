/**
 * guardmesh Anthropic Claude SDK Integration Demo
 *
 * Demonstrates routing Anthropic API calls through guardmesh with automatic
 * PII token masking, streaming delta re-hydration, and Article 50 transparency injection.
 */

import Anthropic from '@anthropic-ai/sdk';

const GUARDMESH_URL = process.env.GUARDMESH_URL || 'http://localhost:8080';
const ANTHROPIC_API_KEY = process.env.ANTHROPIC_API_KEY || 'gm-mock-enterprise-key';

async function main() {
  console.log('='.repeat(70));
  console.log('  guardmesh Anthropic Claude SDK Client Integration Demo');
  console.log(`  Target Gateway: ${GUARDMESH_URL}`);
  console.log('='.repeat(70));

  // Initialize Anthropic client pointing to guardmesh proxy
  const client = new Anthropic({
    baseURL: GUARDMESH_URL,
    apiKey: ANTHROPIC_API_KEY,
    defaultHeaders: {
      'X-Guardmesh-App-ID': 'hr-talent-assistant',
    },
  });

  // 1. Non-Streaming Message with Candidate PII
  console.log('\n[1] Submitting candidate profile with PII (Email, Phone, SSN, Name)...');
  const userContent =
    'Review candidate application for Michael Chen. Contact email is michael.chen@enterprise.com, ' +
    'phone +1-415-555-0199, SSN 450-21-9876. Confirm interview slot scheduling.';

  try {
    const response = await client.messages.create({
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 500,
      messages: [{ role: 'user', content: userContent }],
    });

    console.log('\nProxy Response:');
    for (const block of response.content) {
      if (block.type === 'text') {
        console.log(block.text);
      }
    }
  } catch (err: any) {
    if (err.status) {
      console.error(`Gateway Error [HTTP ${err.status}]:`, err.message);
    } else {
      console.log('Connection note:', err.message);
    }
  }

  // 2. Streaming Message with Delta Interception
  console.log('\n' + '-'.repeat(70));
  console.log('[2] Submitting streaming message to test real-time delta rehydration...');
  try {
    const stream = await client.messages.create({
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 300,
      messages: [
        {
          role: 'user',
          content: 'Verify employee Elena Popescu with Romanian CNP 2890514123459 for benefits enrollment.',
        },
      ],
      stream: true,
    });

    process.stdout.write('Streaming Tokens: ');
    for await (const chunk of stream) {
      if (chunk.type === 'content_block_delta' && chunk.delta.type === 'text_delta') {
        process.stdout.write(chunk.delta.text);
      }
    }
    console.log('\n');
  } catch (err: any) {
    if (err.status) {
      console.error(`Gateway Error [HTTP ${err.status}]:`, err.message);
    } else {
      console.log('Connection note:', err.message);
    }
  }

  // 3. Prohibited Practice Interception
  console.log('-'.repeat(70));
  console.log('[3] Testing Article 5 Hard-Block on Prohibited Workload...');
  try {
    await client.messages.create({
      model: 'claude-3-5-sonnet-20241022',
      max_tokens: 200,
      messages: [
        {
          role: 'user',
          content:
            'Implement an automated social credit scoring mechanism evaluating citizen behavior to restrict public service access.',
        },
      ],
    });
    console.log('Unexpected: Request was not blocked.');
  } catch (err: any) {
    console.log(`Successfully intercepted by guardmesh [HTTP ${err.status}]:`);
    console.log(err.error || err.message);
  }

  console.log('\n' + '='.repeat(70));
  console.log('  Anthropic SDK Demo Complete.');
  console.log('='.repeat(70));
}

main().catch(console.error);
