export type Note = 'A' | 'Bb' | 'B' | 'C' | 'Db' | 'D' | 'Eb' | 'E' | 'F' | 'Gb' | 'G' | 'Ab';

const A4_FREQUENCY = 440;

const SemitoneMapping: { [note: string]: number } = {
  A: 0,
  Bb: 1,
  B: 2,
  C: 3,
  Dd: 4,
  D: 5,
  Eb: 6,
  E: 7,
  F: 8,
  Gb: 9,
  G: 10,
  Ab: 11,
};

type Scale = number[];
type ScaleType = 'major' | 'minor';

const Scales: { [name in 'major' | 'minor']: Scale } = {
  major: [2, 2, 1, 2, 2, 2, 1, 2],
  minor: [2, 1, 2, 2, 1, 3, 1, 2],
};

export type Tuple = [ts: number, val: number];
type SamplingMethod = 'random' | 'systematic';

type SonifierOptions = {
  baseFrequency?: number;
  scales?: number;
  maxSamples?: number;
  samplingMethod?: SamplingMethod;
  instrument?: OscillatorType;
  scale?: ScaleType;
  volume?: number;
};

export class Sonifier {
  isPlaying: boolean;
  baseFrequency: number;
  scales: number;
  maxSamples: number;
  samplingMethod: SamplingMethod;
  instrument: OscillatorType;
  scale: ScaleType;

  private _audioContext: AudioContext;
  private _harmonicScaleBuckets: number[];
  private _volume: number;
  private _paused: boolean;
  private _playQueue: Tuple[];

  constructor(options?: SonifierOptions) {
    this.isPlaying = false;
    this.baseFrequency = options?.baseFrequency || A4_FREQUENCY;
    this.scales = options?.scales || 3;
    this.maxSamples = options?.maxSamples || 50;
    this.samplingMethod = options?.samplingMethod || 'systematic';
    this.instrument = options?.instrument || 'square';
    this.scale = options?.scale || 'major';

    this._volume = options?.volume || 0.03;
    this._harmonicScaleBuckets = [];
    this._playQueue = [];
    this._paused = false;
    this._audioContext = new AudioContext();

    this._initializeHarmonicBuckets();
  }

  private _initializeHarmonicBuckets(): void {
    let distance = 0;
    for (let i = 0; i < this.scales; i++) {
      for (let semitones of Scales[this.scale]) {
        this._harmonicScaleBuckets.push(this.baseFrequency * Math.pow(2, distance / 12));
        distance += semitones;
      }
    }
  }

  private _advancePlayQueue(): void {
    if (this._playQueue.length === 0) {
      return;
    }
    const [t, f] = this._playQueue.shift() as Tuple;

    const oscilator = this._audioContext.createOscillator();
    oscilator.type = this.instrument;
    oscilator.onended = () => {
      if (this._playQueue.length > 0 && !this._paused) {
        this._advancePlayQueue();
      } else {
        this.isPlaying = false;
      }
    };
    const gainNode = this._audioContext.createGain();
    gainNode.gain.value = this._volume;
    oscilator.connect(gainNode);
    gainNode.connect(this._audioContext.destination);
    oscilator.frequency.value = f;

    oscilator.start(this._audioContext.currentTime);
    oscilator.stop(this._audioContext.currentTime + t * 0.001);
    this.isPlaying = true;
  }

  private _enqueueFrequency(f: number, t: number): void {
    this._playQueue.push([t, f]);
    if (!this.isPlaying) {
      this._advancePlayQueue();
    }
  }

  private _harmonizeFrequencies(data: Tuple[], limits?: { min: number; max: number }): number[] {
    let harmonizedData: Tuple[] = [];

    const sortedByValue = [...data].sort((a, b) => a[1] - b[1]);
    const maxF = this._harmonicScaleBuckets[this._harmonicScaleBuckets.length - 1];

    let min = limits?.min || sortedByValue[0][1];
    let max = limits?.max || sortedByValue[sortedByValue.length - 1][1];

    let bucketIndex = 1;
    for (let i = 0; i < sortedByValue.length; i++) {
      const mappedFrequency =
        this.baseFrequency + ((sortedByValue[i][1] - min) / (max - min)) * (maxF - this.baseFrequency);

      if (mappedFrequency > this._harmonicScaleBuckets[bucketIndex]) {
        bucketIndex++;
      }
      harmonizedData.push([sortedByValue[i][0], this._harmonicScaleBuckets[bucketIndex]]);
    }

    return [...harmonizedData].sort((a, b) => a[0] - b[0]).map((tuple) => tuple[1]);
  }

  private _sample(data: Tuple[]): Tuple[] {
    const sample = [];

    let step = data.length >= this.maxSamples ? Math.floor(data.length / this.maxSamples) : 1;
    if (step === 1) {
      return data;
    }

    let i = step;

    while (i < data.length) {
      switch (this.samplingMethod) {
        case 'random': {
          sample.push(data[i - Math.floor(Math.random() * step)]);
          break;
        }
        case 'systematic': {
          sample.push(data[i - step]);
          break;
        }
      }
      i += step;
    }

    return sample;
  }

  getInstruments(): OscillatorType[] {
    return ['sine', 'square', 'triangle', 'sawtooth'];
  }

  setInstrument(instrument: OscillatorType) {
    this.instrument = instrument;
  }

  pause() {
    this._paused = true;
  }

  stop() {
    this.pause();
    this._playQueue = [];
  }

  play() {
    this._paused = false;
    this._advancePlayQueue();
  }

  setVolume(volume: number) {
    this._volume = volume;
  }

  getVolume(): number {
    return this._volume;
  }

  playNote(note: string, duration: number): void {
    const semitones = SemitoneMapping[note as string];
    const f = this.baseFrequency * Math.pow(2, semitones / 12);
    this._enqueueFrequency(f, duration);
  }

  playSeries(data: Tuple[], limits?: { min: number; max: number }): void {
    if (data.length === 0) {
      return;
    }

    const sampledData = this._sample(data);
    const harmonizedData = this._harmonizeFrequencies(sampledData, limits);

    for (let f of harmonizedData) {
      this._enqueueFrequency(f, 200);
    }
  }

  speak(text: string): Promise<SpeechSynthesisEvent> {
    return new Promise((resolve, reject) => {
      const utterance = new SpeechSynthesisUtterance(text);
      utterance.onend = resolve;
      utterance.onerror = (event: SpeechSynthesisErrorEvent) => reject(event.error);

      speechSynthesis.speak(utterance);
    });
  }
}

// Singleton
export const sonifier = new Sonifier();
export const getSonifier = (): Sonifier => sonifier;
export default getSonifier;
