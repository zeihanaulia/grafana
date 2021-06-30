async function sleep(ms: number) {
  return new Promise((resolve) => {
    setTimeout(resolve, ms);
  });
}

const A4_FREQUENCY = 440;

type Scale = number[];
type ScaleType = 'major' | 'minor';

const Scales: { [name in 'major' | 'minor']: Scale } = {
  major: [2, 2, 1, 2, 2, 2, 1, 2],
  minor: [2, 1, 2, 2, 1, 3, 1, 2],
};

type Tuple = [ts: number, val: number];
type SamplingMethod = 'random' | 'systematic';

type SonifierOptions = {
  baseFrequency?: number;
  scales?: number;
  maxSamples?: number;
  samplingMethod?: SamplingMethod;
  instrument?: OscillatorType;
  scale?: ScaleType;
};

class Sonifier {
  isPlaying: boolean;
  baseFrequency: number;
  scales: number;
  maxSamples: number;
  samplingMethod: SamplingMethod;
  instrument: OscillatorType;
  scale: ScaleType;

  private _audioContext: AudioContext;
  private _harmonicScaleBuckets: number[];

  constructor(options: SonifierOptions) {
    this.isPlaying = false;
    this.baseFrequency = options.baseFrequency || A4_FREQUENCY;
    this.scales = options.scales || 3;
    this.maxSamples = options.maxSamples || 50;
    this.samplingMethod = options.samplingMethod || 'systematic';
    this.instrument = options.instrument || 'square';
    this.scale = options.scale || 'major';

    this._harmonicScaleBuckets = [];
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

  private _playFrequency(f: number, t: number): void {
    const oscilator = this._audioContext.createOscillator();
    oscilator.type = this.instrument;
    oscilator.frequency.value = f;

    const gainNode = this._audioContext.createGain();

    oscilator.connect(gainNode);
    gainNode.connect(this._audioContext.destination);

    oscilator.start(0);
    this.isPlaying = true;

    setTimeout(() => {
      oscilator.stop();
      this.isPlaying = false;
    }, t);
  }

  private _harmonizeFrequencies(data: Tuple[]): number[] {
    let harmonizedData: Tuple[] = [];

    const sortedByValue = [...data].sort((a, b) => a[1] - b[1]);
    const maxF = this._harmonicScaleBuckets[this._harmonicScaleBuckets.length - 1];

    let min = sortedByValue[0][1],
      max = sortedByValue[sortedByValue.length - 1][1];

    let bucketIndex = 1;
    for (let i = 0; i < sortedByValue.length; i++) {
      const mappedFrequency =
        this.baseFrequency + ((sortedByValue[i][1] - min) / (max - min)) * (maxF - this.baseFrequency);

      if (mappedFrequency >= this._harmonicScaleBuckets[bucketIndex]) {
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

  async playSeries(data: Tuple[]): Promise<void> {
    if (data.length === 0) {
      return;
    }

    const sampledData = this._sample(data);
    const harmonizedData = this._harmonizeFrequencies(sampledData);

    for (let f of harmonizedData) {
      await sleep(220);
      this._playFrequency(f, 200);
    }
  }

  speak(text: string): Promise<void> {
    const utter = new SpeechSynthesisUtterance(text);
    speechSynthesis.speak(utter);

    return Promise.resolve();
  }
}

export default Sonifier;
